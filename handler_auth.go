package main

import (
	"encoding/json"
	"net/http"
	"time"

	"boot.dev/chirpy/internal/auth"
	"boot.dev/chirpy/internal/database"
)

func (cfg *apiConfig) handlerAuthLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Couldn't find user with that email.",
			err,
		)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword.String)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect password", err)
		return
	}

	token, err := auth.MakeJWT(
		user.ID,
		cfg.secret,
		time.Duration(3600)*time.Second, // One hour
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to generate JWT", err)
		return
	}

	rTokenStr, err := auth.MakeRefreshToken()
	rToken, err := cfg.db.CreateRefreshToken(
		r.Context(),
		database.CreateRefreshTokenParams{Token: rTokenStr, UserID: user.ID},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to generate Refresh Token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        token,
		RefreshToken: rToken.Token,
	})
}

func (cfg *apiConfig) handlerAuthRefresh(w http.ResponseWriter, r *http.Request) {
	type parameters struct{}
	type response struct {
		Token string `json:"token"`
	}

	rTokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "No Bearer token provided in header.", err)
		return
	}
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), rTokenStr)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token.", err)
		return
	}
	token, err := auth.MakeJWT(
		user.ID,
		cfg.secret,
		time.Duration(3600)*time.Second, // One hour
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to generate JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: token,
	})
}

func (cfg *apiConfig) handlerAuthRevoke(w http.ResponseWriter, r *http.Request) {
	type parameters struct{}
	type response struct{}

	rTokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "No Bearer token provided in header.", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), rTokenStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid refresh token.", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
  return
}
