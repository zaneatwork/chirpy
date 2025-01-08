package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"boot.dev/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func main() {
	const filepathRoot = "."
	const port = "8080"

  godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))),
	)
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerReset)
	// mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	// mux.HandleFunc("GET /api/chirps/", apiCfg.handlerChirpRetrieve)
	// mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsRetrieve)
	// mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	// mux.HandleFunc("GET /api/users/", apiCfg.handlerUserRetrieve)
	// mux.HandleFunc("GET /api/users", apiCfg.handlerUsersRetrieve)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
