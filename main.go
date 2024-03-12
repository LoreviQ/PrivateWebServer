package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type apiConfig struct {
	port           string
	dbDirectory    string
	jwtSecret      []byte
	fileserverHits int
	db             Database
}

func initialiseServer(cfg apiConfig, mux *http.ServeMux) *http.Server {
	const filepathRoot = "."

	mux.Handle("/app/*", http.StripPrefix("/app", cfg.metricsIncMiddleware(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsReportingHandler)
	mux.HandleFunc("GET /api/reset", cfg.metricsResetHandler)
	mux.HandleFunc("GET /api/chirps", cfg.getChirpHandler)
	mux.HandleFunc("POST /api/chirps", cfg.postChirpHandler)
	mux.HandleFunc("GET /api/chirps/{id}", cfg.getChirpByIDHandler)
	mux.HandleFunc("POST /api/users", cfg.postUserHandler)
	mux.HandleFunc("POST /api/login", cfg.postLoginHandler)

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
	godotenv.Load()
	cfg := apiConfig{
		port:           "8080",
		dbDirectory:    "./database/database.json",
		jwtSecret:      []byte(os.Getenv("JWT_SECRET")),
		fileserverHits: 0,
	}
	cfg.handleFlags()
	cfg.db = initialiseDatabase(cfg.dbDirectory)
	mux := http.NewServeMux()
	server := initialiseServer(cfg, mux)

	log.Printf("Serving on port: %s\n", cfg.port)
	log.Panic(server.ListenAndServe())
}
