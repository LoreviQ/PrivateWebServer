package main

import (
	"log"
	"net/http"
	"os"

	"github.com/LoreviQ/PrivateWebServer/internal/db"
	"github.com/LoreviQ/PrivateWebServer/internal/hdl"
	"github.com/joho/godotenv"
)

func initialiseServer(cfg hdl.ApiConfig, mux *http.ServeMux) *http.Server {
	const filepathRoot = "."

	mux.Handle("/app/*", http.StripPrefix("/app", cfg.MetricsIncMiddleware(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", cfg.HealthzHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.MetricsReportingHandler)
	mux.HandleFunc("GET /api/reset", cfg.MetricsResetHandler)
	mux.HandleFunc("GET /api/chirps", cfg.GetChirpHandler)
	mux.HandleFunc("POST /api/chirps", cfg.PostChirpHandler)
	mux.HandleFunc("GET /api/chirps/{id}", cfg.GetChirpByIDHandler)
	mux.HandleFunc("POST /api/users", cfg.PostUserHandler)
	mux.HandleFunc("PUT /api/users", cfg.PutUserHandler)
	mux.HandleFunc("POST /api/login", cfg.PostLoginHandler)
	mux.HandleFunc("POST /api/refresh", cfg.PostRefreshHandler)
	mux.HandleFunc("POST /api/revoke", cfg.PostRevokeHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.DeleteChirpHandler)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.PostPolkaWebhook)

	corsMux := cfg.CorsMiddleware(mux)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: corsMux,
	}
	return server
}

func main() {
	godotenv.Load()
	cfg := hdl.ApiConfig{
		Port:           "8080",
		DB_Directory:   "./database/database.json",
		JWT_Secret:     []byte(os.Getenv("JWT_SECRET")),
		FileserverHits: 0,
	}
	cfg.HandleFlags()
	cfg.DB = db.InitialiseDatabase(cfg.DB_Directory)
	mux := http.NewServeMux()
	server := initialiseServer(cfg, mux)

	log.Printf("Serving on port: %s\n", cfg.Port)
	log.Panic(server.ListenAndServe())
}
