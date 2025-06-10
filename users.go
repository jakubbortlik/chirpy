package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jakubbortlik/chirpy/internal/auth"
	"github.com/jakubbortlik/chirpy/internal/database"
)

type User struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     *string   `json:"email"`
}

func handlerUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
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

	hashedPassword, err := auth.HashPassword(*params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Connecting to database failed", err)
		return
	}

	dbQueries := database.New(db)
	createUserParams := database.CreateUserParams{
		Email:          *params.Email,
		HashedPassword: hashedPassword,
	}
	user, err := dbQueries.CreateUser(r.Context(), createUserParams)

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
