package main

import (
	"database/sql"
	"encoding/json"
	"github.com/jakubbortlik/chirpy/internal/database"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password *string `json:"password"`
		Email    *string `json:"email"`
	}
	type response struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
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
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Connecting to database failed", err)
		return
	}
	dbQueries := database.New(db)
	user, err := dbQueries.GetUser(r.Context(), *params.Email)
	errCompare := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(*params.Password))
	if err != nil || errCompare != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Id: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	})
}
