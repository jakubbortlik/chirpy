package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/jakubbortlik/chirpy/internal/auth"
	"github.com/jakubbortlik/chirpy/internal/database"
)

func handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No token in header", err)
		return
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Connecting to database failed", err)
		return
	}

	dbQueries := database.New(db)
	err = dbQueries.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		Token:     token,
		UpdatedAt: time.Now(),
	})

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't revoke token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
