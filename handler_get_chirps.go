package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/jakubbortlik/chirpy/internal/database"
)

func handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Opening database connection failed.", err)
		return
	}
	dbQueries := database.New(db)
	chirps_data, err := dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Getting chirps from database failed.", err)
		return
	}
	var chirps []Chirp
	for _, chirp := range chirps_data {
		chirps = append(chirps, Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
	return
}

func handlerGetIndividualChirp(w http.ResponseWriter, r *http.Request) {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Opening database connection failed.", err)
		return
	}
	dbQueries := database.New(db)
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Parsing chirpID failed.", err)
		return
	}
	chirp_data, err := dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Getting chirps from database failed.", err)
		return
	}
	chirp := Chirp{
		Id:        chirp_data.ID,
		CreatedAt: chirp_data.CreatedAt,
		UpdatedAt: chirp_data.UpdatedAt,
		Body:      chirp_data.Body,
		UserID:    chirp_data.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirp)
	return
}
