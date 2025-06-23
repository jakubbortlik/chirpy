package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/jakubbortlik/chirpy/internal/auth"
	"github.com/jakubbortlik/chirpy/internal/database"
)

func handlerRefreshToken(w http.ResponseWriter, r *http.Request, apiConfig *apiConfig) {
	type response struct {
		Token string `json:"token"`
	}

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
	refreshToken, err := dbQueries.GetRefreshToken(r.Context(), token)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed getting token from database", err)
		return
	}

	if refreshToken.ExpiresAt.Before(time.Now())  {
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired", err)
		return
	}

	if refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token revoked", err)
		return
	}

	userID, err := dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user ID from database", err)
		return
	}

	JWTToken, err := auth.MakeJWT(
		userID,
		apiConfig.JWTSecret,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT token", err)
	}
	
	respondWithJSON(w, http.StatusOK, response{
		Token: JWTToken,
	})
}
