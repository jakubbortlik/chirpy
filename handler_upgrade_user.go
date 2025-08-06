package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/jakubbortlik/chirpy/internal/database"
)

func handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		UserID uuid.UUID `json:"user_id"`
	}
	type parameters struct {
		Event *string `json:"event"`
		Data  Data    `json:"data"`
	}
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Decoding parameters failed", err)
		return
	}

	if *params.Event != "user.upgraded" {
		respondWithNoBody(w, http.StatusNoContent)
		return
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Connecting to database failed", err)
		return
	}

	dbQueries := database.New(db)
	_, err = dbQueries.UpgradeUser(r.Context(), params.Data.UserID)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "User not found in database", err)
		return
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Upgrading user failed", err)
		return
	}

	respondWithNoBody(w, http.StatusNoContent)
}
