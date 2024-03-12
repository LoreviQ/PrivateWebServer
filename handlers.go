package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
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

func (db *Database) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirps := make([]Chirp, len(db.Chirps))
	for i, chirp := range db.Chirps {
		chirps[i-1] = chirp
	}
	data, err := json.Marshal(chirps)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (db *Database) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, 400, "Invalid ID")
		return
	}
	chirp, ok := db.Chirps[id]
	if !ok {
		writeError(w, 404, "No Chirp by that ID")
		return
	}
	data, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (db *Database) postChirpHandler(w http.ResponseWriter, r *http.Request) {
	type requestStruct struct {
		Body string `json:"body"`
	}
	request := requestStruct{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if len(request.Body) <= 140 {
		db.writeValidatedChirp(w, request.Body)
	} else {
		writeError(w, 400, "Chirp is too long")
	}
}

func (db *Database) postUserHandler(w http.ResponseWriter, r *http.Request) {
	type requestStruct struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	request := requestStruct{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		log.Printf("Error Generating password hash: %s", err)
		w.WriteHeader(500)
		return
	}
	user, err := db.addUser(request.Email, hash)
	if errors.Is(err, ErrTakenEmail) {
		writeError(w, 400, "This email has already been taken")
		return
	} else if err != nil {
		log.Printf("Error Adding user: %s", err)
		w.WriteHeader(500)
		return
	}

	type Response struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(Response{
		Email: user.Email,
		ID:    user.ID,
	})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)
	w.Write(data)
}

func (db *Database) postLoginHandler(w http.ResponseWriter, r *http.Request) {
	type requestStruct struct {
		Password string `json:"password"`
		Email    string `json:"email"`
		Timeout  int    `json:"expires_in_seconds"`
	}
	request := requestStruct{}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := db.authenticateUser(request.Email, []byte(request.Password))
	if errors.Is(err, ErrInvalidEmail) {
		writeError(w, 404, "No user with this email")
		return
	} else if errors.Is(err, ErrIncorrectPassword) {
		writeError(w, 401, "Incorrect Password")
		return
	} else if err != nil {
		log.Printf("Error Authenticating User: %s", err)
		w.WriteHeader(500)
		return
	}

	type Response struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(Response{
		Email: user.Email,
		ID:    user.ID,
	})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}

func (db *Database) writeValidatedChirp(w http.ResponseWriter, body string) {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(body, " ")
	for i, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	chirp, err := db.createChirp(strings.Join(words, " "))
	if err != nil {
		log.Printf("Error creating Chirp: %s", err)
	}

	data, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)
	w.Write(data)
}

func writeError(w http.ResponseWriter, errorCode int, errorText string) {
	type ReturnVals struct {
		Error string `json:"error"`
	}
	data, err := json.Marshal(ReturnVals{Error: errorText})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(errorCode)
	w.Write(data)
}
