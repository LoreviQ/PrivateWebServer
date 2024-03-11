package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	db := initialiseDatabase("./database/database.json")
	mux := http.NewServeMux()
	cfg := apiConfig{fileserverHits: 0}
	fmt.Print(db)

	mux.Handle("/app/*", http.StripPrefix("/app", cfg.metricsIncMiddleware(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsReportingHandler)
	mux.HandleFunc("GET /api/reset", cfg.metricsResetHandler)
	mux.HandleFunc("POST /api/chirps", postChirpHandler)

	corsMux := corsMiddleware(mux)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
