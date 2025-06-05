package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jakubbortlik/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string   `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func handlerPostChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   *string   `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	type response struct {
		Chirp
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Decoding parameters failed", err)
		return
	}

	cleanedBody, err := validateChirp(params.Body)

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	var chirp database.Chirp
	if err == nil {
		dbQueries := database.New(db)
		chirp, err = dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   cleanedBody,
			UserID: params.UserId,
		})
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Creating chirp failed.", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		Chirp: Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      cleanedBody,
			UserID:    params.UserId,
		},
	})
	return
}

func validateChirp(body *string) (string, error) {
	if body == nil {
		return "", errors.New("The request body doesn't contain the required field `body`")
	}
	const maxChirpLength = 140
	if chirp_length := len(*body); chirp_length > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}
	const profanityReplacement = "****"
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleanedBody := getCleanedbody(body, badWords, profanityReplacement)
	return cleanedBody, nil
}
func getCleanedbody(body *string, badWords map[string]struct{}, replacement string) string {
	words := strings.Split(*body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = replacement
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
