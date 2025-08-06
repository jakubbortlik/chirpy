package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/jakubbortlik/chirpy/internal/auth"
	"github.com/jakubbortlik/chirpy/internal/database"
)

func handlerDeleteChirp(w http.ResponseWriter, r *http.Request, apiConfig *apiConfig) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No token in header", err)
		return
	}

	userID, err := auth.ValidateJWT(token, apiConfig.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Decoding parameters failed", err)
		return
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Parsing chirpID failed.", err)
		return
	}

	chirp, err := dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found.", err)
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Not allowed to delete chirp.", err)
		return
	}

	err = dbQueries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found.", err)
		return
	}

	respondWithNoBody(w, http.StatusNoContent)
}
