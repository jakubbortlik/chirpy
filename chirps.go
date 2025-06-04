package main

import "github.com/jakubbortlik/chirpy/internal/database"
import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

func handlerChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body *string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	type returnVals struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body *string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if params.Body == nil {
		respondWithError(w, http.StatusBadRequest, "The request body doesn't contain the required field `body`", nil)
		return
	}
	const maxChirpLength = 140
	if chirp_length := len(*params.Body); chirp_length > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	const profanityReplacement = "****"
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	cleanedBody := getCleanedbody(params.Body, badWords, profanityReplacement)

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	var chirp database.Chirp
	if err == nil {
		dbQueries := database.New(db)
		chirp, err = dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
			Body: *cleanedBody,
			UserID: params.UserId,
		})
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Creating chirp failed.", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, returnVals{
		Id: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: cleanedBody,
		UserId: params.UserId,
	})
	return
}

func getCleanedbody(body *string, badWords map[string]struct{}, replacement string) *string {
	words := strings.Split(*body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = replacement
		}
	}
	cleaned := strings.Join(words, " ")
	return &cleaned
}
