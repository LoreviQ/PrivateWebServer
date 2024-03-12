package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

type apiConfig struct {
	port           string
	dbDirectory    string
	fileserverHits int
}

func initialiseServer(cfg apiConfig, db Database, mux *http.ServeMux) *http.Server {
	const filepathRoot = "."

	mux.Handle("/app/*", http.StripPrefix("/app", cfg.metricsIncMiddleware(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsReportingHandler)
	mux.HandleFunc("GET /api/reset", cfg.metricsResetHandler)
	mux.HandleFunc("GET /api/chirps", db.getChirpHandler)
	mux.HandleFunc("POST /api/chirps", db.postChirpHandler)
	mux.HandleFunc("GET /api/chirps/{id}", db.getChirpByIDHandler)
	mux.HandleFunc("POST /api/users", db.postUserHandler)
	mux.HandleFunc("POST /api/login", db.postLoginHandler)

	corsMux := corsMiddleware(mux)

	server := &http.Server{
		Addr:    ":" + cfg.port,
		Handler: corsMux,
	}
	return server
}

func (cfg *apiConfig) handleFlags() {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		log.Printf("Entering debug mode\n")
		cfg.dbDirectory = "./database/debugDB.json"
		os.Remove(cfg.dbDirectory)
	}
}

func main() {
	cfg := apiConfig{
		port:           "8080",
		dbDirectory:    "./database/database.json",
		fileserverHits: 0,
	}
	cfg.handleFlags()
	db := initialiseDatabase(cfg.dbDirectory)
	mux := http.NewServeMux()
	server := initialiseServer(cfg, db, mux)

	log.Printf("Serving on port: %s\n", cfg.port)
	log.Panic(server.ListenAndServe())
}
