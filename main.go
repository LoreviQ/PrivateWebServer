package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	db := initialiseDatabase("./database/database.json")
	mux := http.NewServeMux()
	cfg := apiConfig{fileserverHits: 0}

	mux.Handle("/app/*", http.StripPrefix("/app", cfg.metricsIncMiddleware(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsReportingHandler)
	mux.HandleFunc("GET /api/reset", cfg.metricsResetHandler)
	mux.HandleFunc("GET /api/chirps", db.getChirpHandler)
	mux.HandleFunc("POST /api/chirps", db.postChirpHandler)
	mux.HandleFunc("GET /api/chirps/{id}", db.getChirpByIDHandler)

	corsMux := corsMiddleware(mux)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
