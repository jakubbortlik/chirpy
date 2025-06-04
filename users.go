package main

import "github.com/jakubbortlik/chirpy/internal/database"
import (
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"os"
	"time"
)

type User struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     *string   `json:"email"`
}

func handlerUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email *string `json:"email"`
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

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	var user database.User
	if err == nil {
		dbQueries := database.New(db)
		user, err = dbQueries.CreateUser(r.Context(), *params.Email)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Creating user failed", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			Id:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     &user.Email,
		},
	})
}
