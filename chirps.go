package main

import (
	"encoding/json"
	"net/http"
	"regexp"
)

func handlerChirpValidation(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body *string `json:"body"`
	}
	type returnVals struct {
		CleanedBody *string `json:"cleaned_body"`
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
	re := regexp.MustCompile(`(?i)(kerfuffle|sharbert|fornax)`)
	cleaned_body := re.ReplaceAllString(*params.Body, profanityReplacement)
	respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: &cleaned_body,
	})
	return
}
