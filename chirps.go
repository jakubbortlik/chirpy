package main

import (
	"encoding/json"
	"net/http"
	"strings"
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
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	cleanedBody := getCleanedbody(params.Body, badWords, profanityReplacement)
	respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: cleanedBody,
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
