package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Status", "200 OK")
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsReportingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Status", "200 OK")
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %v times!</p></body></html>", cfg.fileserverHits)))
}

func (cfg *apiConfig) metricsResetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Status", "200 OK")
	cfg.fileserverHits = 0
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		Body string `json:"body"`
	}

	chirp := Chirp{}
	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(chirp.Body) <= 140 {
		writeValidated(w)
	} else {
		writeError(w)
	}
}

func writeValidated(w http.ResponseWriter) {
	type ReturnVals struct {
		Valid bool `json:"valid"`
	}
	data, err := json.Marshal(ReturnVals{Valid: true})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}

func writeError(w http.ResponseWriter) {
	type ReturnVals struct {
		Error string `json:"error"`
	}
	data, err := json.Marshal(ReturnVals{Error: "Chirp is too long"})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(400)
	w.Write(data)
}
