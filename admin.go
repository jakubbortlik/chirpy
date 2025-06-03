package main

import "github.com/jakubbortlik/chirpy/internal/database"
import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if platform := os.Getenv("PLATFORM"); platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Resetting only allowed in local dev environment.", nil)
		return
	}
	cfg.fileserverHits.Store(0)

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err == nil {
		dbQueries := database.New(db)
		err = dbQueries.DeleteUsers(r.Context())
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Deleting users failed.", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Deleted all users. Hits reset to 0."))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	html := fmt.Sprintf(`
<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
</html>
	`, cfg.fileserverHits.Load(),
	)
	w.Write([]byte(html))
}
