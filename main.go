package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	fs := http.FileServer(http.Dir(filepathRoot))

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", fs))
	mux.HandleFunc("/healthz", handlerReadiness)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
