package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/jakubbortlik/chirpy/internal/auth"
	"github.com/jakubbortlik/chirpy/internal/database"
)

func handlerUpdateUser(w http.ResponseWriter, r *http.Request, apiConfig *apiConfig) {
	type parameters struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	type response struct {
		User
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No token in header", err)
		return
	}

	userID, err := auth.ValidateJWT(token, apiConfig.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
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
	updateUserParams := database.UpdateUserParams{
		ID:             userID,
		Email:          *params.Email,
		HashedPassword: hashedPassword,
	}
	user, err := dbQueries.UpdateUser(r.Context(), updateUserParams)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Updating user failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			Id:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       &user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}
