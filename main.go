package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	cfg := apiConfig{fileserverHits: 0}
	mux.Handle("/app/*", http.StripPrefix("/app", cfg.metricsIncMiddleware(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/metrics", cfg.metricsReportingHandler)

	corsMux := corsMiddleware(mux)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
