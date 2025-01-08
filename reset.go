package main

import (
	"net/http"
	"sync/atomic"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = atomic.Int32{}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
