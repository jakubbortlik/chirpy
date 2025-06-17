package main

import (
	"database/sql"
	"encoding/json"
	"github.com/jakubbortlik/chirpy/internal/auth"
	"github.com/jakubbortlik/chirpy/internal/database"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func handlerUserLogin(w http.ResponseWriter, r *http.Request, apiConfig *apiConfig) {
	type parameters struct {
		Password         *string `json:"password"`
		Email            *string `json:"email"`
		ExpiresInSeconds *int32  `json:"expires_in_seconds,omitempty"`
	}
	type response struct {
		User
		Token string `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Decoding parameters failed", err)
		return
	}

	var expirationTime int32
	if params.ExpiresInSeconds == nil || *params.ExpiresInSeconds > 3600 {
		expirationTime = 3600
	} else {
		expirationTime = *params.ExpiresInSeconds
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

	JWTToken, err := auth.MakeJWT(
		user.ID,
		apiConfig.JWTSecret,
		time.Duration(expirationTime)*time.Second,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create token", err)
	}
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			Id:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     &user.Email,
		},
		Token: JWTToken,
	})
}
