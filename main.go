package main

import (
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	JWTSecret      string
	PolkaKey       string
}

func main() {
	godotenv.Load()

	const filepathRoot = "."
	const port = "8080"

	fs := http.FileServer(http.Dir(filepathRoot))
	apiCfg := &apiConfig{
		JWTSecret: os.Getenv("JWT_SECRET"),
		PolkaKey:  os.Getenv("POLKA_KEY"),
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fs)))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		handlerPostChirp(w, r, apiCfg)
	})
	mux.HandleFunc("GET /api/chirps", handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", handlerGetIndividualChirp)

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
		handlerDeleteChirp(w, r, apiCfg)
	})

	mux.HandleFunc("POST /api/users", handlerCreateUser)
	mux.HandleFunc("PUT /api/users", func(w http.ResponseWriter, r *http.Request) {
		handlerUpdateUser(w, r, apiCfg)
	})

	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		handlerUserLogin(w, r, apiCfg)
	})
	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		handlerRefreshToken(w, r, apiCfg)
	})
	mux.HandleFunc("POST /api/revoke", handlerRevokeToken)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("POST /api/polka/webhooks", func(w http.ResponseWriter, r *http.Request) {
		handlerUpgradeUser(w, r, apiCfg)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
