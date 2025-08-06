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
		Password *string `json:"password"`
		Email    *string `json:"email"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	refreshToken := auth.MakeRefreshToken()
	expiresAt := time.Now().Add(60 * 24 * time.Hour)

	createTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}

	err = dbQueries.CreateRefreshToken(r.Context(), createTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Creating refresh token failed", err)
	}

	JWTToken, err := auth.MakeJWT(
		user.ID,
		apiConfig.JWTSecret,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT token", err)
	}
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			Id:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       &user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        JWTToken,
		RefreshToken: refreshToken,
	})
}
