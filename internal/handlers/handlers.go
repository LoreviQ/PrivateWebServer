package hdl

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/LoreviQ/PrivateWebServer/internal/auth"
	"github.com/LoreviQ/PrivateWebServer/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type apiConfig struct {
	port           string
	dbDirectory    string
	jwtSecret      []byte
	fileserverHits int
	db             db.Database
}

func (cfg *apiConfig) healthzHandler(w http.ResponseWriter, r *http.Request) {
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

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirps := make([]db.Chirp, len(cfg.db.Chirps))
	for i, chirp := range cfg.db.Chirps {
		chirps[i-1] = chirp
	}
	writeResponse(w, 200, chirps)
}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, 400, "Invalid ID")
		return
	}
	chirp, ok := cfg.db.Chirps[id]
	if !ok {
		writeError(w, 404, "No Chirp by that ID")
		return
	}
	writeResponse(w, 200, chirp)
}

func (cfg *apiConfig) postChirpHandler(w http.ResponseWriter, r *http.Request) {
	// CHECKING AUTHENTICATION
	id, err := auth.AuthenticateAccessToken(r, cfg.jwtSecret)
	if err != nil {
		writeError(w, 401, "Inavlid Token. Please log in again")
		return
	}

	// REQUEST
	type requestStruct struct {
		Body string `json:"body"`
	}
	request, err := decodeRequest(w, r, requestStruct{})
	if err != nil {
		return
	}

	// FUNCTION BODY
	w.Header().Set("Content-Type", "application/json")
	if len(request.Body) <= 140 {
		chirp, err := cfg.validateChirp(request.Body, id)
		if err != nil {
			log.Printf("Error validating chirp: %s", err)
			w.WriteHeader(500)
			return
		}

		// RESPONSE
		writeResponse(w, 201, chirp)
	} else {
		writeError(w, 400, "Chirp is too long")
	}
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	// AUTHENTICATION
	userID, err := auth.AuthenticateAccessToken(r, cfg.jwtSecret)
	if err != nil {
		writeError(w, 401, "Inavlid Token. Please log in again")
		return
	}

	// AUTHORIZATION
	chirpID, err := strconv.Atoi(r.PathValue("chirpID"))
	if err != nil {
		writeError(w, 400, "Invalid ID")
		return
	}
	chirp, ok := cfg.db.Chirps[chirpID]
	if !ok {
		writeError(w, 404, "No Chirp by that ID")
		return
	}
	if chirp.UserID != userID {
		writeError(w, 403, "Not Authorised to delete this chirp")
		return
	}

	// DELETING CHIRP
	err = cfg.db.DeleteChirp(chirpID)
	if err != nil {
		log.Printf("Error Deleting Chirp: %s", err)
		w.WriteHeader(500)
		return
	}

	// RESPONSE
	w.WriteHeader(200)
}

func (cfg *apiConfig) postUserHandler(w http.ResponseWriter, r *http.Request) {
	// REQUEST
	type requestStruct struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	request, err := decodeRequest(w, r, requestStruct{})
	if err != nil {
		return
	}

	// FUNCTION BODY
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		log.Printf("Error Generating password hash: %s", err)
		w.WriteHeader(500)
		return
	}
	user, err := cfg.db.AddUser(request.Email, hash)
	if errors.Is(err, db.ErrTakenEmail) {
		writeError(w, 400, "This email has already been taken")
		return
	} else if err != nil {
		log.Printf("Error Adding user: %s", err)
		w.WriteHeader(500)
		return
	}

	// RESPONSE
	type responseStruct struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
	}
	writeResponse(w, 201, responseStruct{
		Email: user.Email,
		ID:    user.ID,
	})
}

func (cfg *apiConfig) postLoginHandler(w http.ResponseWriter, r *http.Request) {
	// REQUEST
	type requestStruct struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	request, err := decodeRequest(w, r, requestStruct{})
	if err != nil {
		return
	}

	// AUTHENTICATE USER
	user, err := cfg.db.AuthenticateUser(request.Email, []byte(request.Password))
	if errors.Is(err, db.ErrInvalidEmail) {
		writeError(w, 404, "No user with this email")
		return
	} else if errors.Is(err, db.ErrIncorrectPassword) {
		writeError(w, 401, "Incorrect Password")
		return
	} else if err != nil {
		log.Printf("Error Authenticating User: %s", err)
		w.WriteHeader(500)
		return
	}

	// CREATE JWT TOKENS
	accessToken, err := auth.IssueAccessToken(user.ID, cfg.jwtSecret)
	if err != nil {
		log.Printf("Error Creating Access Token: %s", err)
		w.WriteHeader(500)
		return
	}
	refreshToken, err := auth.IssueRefreshToken(user.ID, cfg.jwtSecret, cfg.db)
	if err != nil {
		log.Printf("Error Creating Refresh Token: %s", err)
		w.WriteHeader(500)
		return
	}

	// RESPONSE
	type responseStruct struct {
		ID           int    `json:"id"`
		Email        string `json:"email"`
		AccessToken  string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	writeResponse(w, 200, responseStruct{
		Email:        user.Email,
		ID:           user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) putUserHandler(w http.ResponseWriter, r *http.Request) {
	// CHECKING AUTHENTICATION
	id, err := auth.AuthenticateAccessToken(r, cfg.jwtSecret)
	if err != nil {
		writeError(w, 401, "Inavlid Token. Please log in again")
		return
	}

	// REQUEST
	type requestStruct struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	request, err := decodeRequest(w, r, requestStruct{})
	if err != nil {
		return
	}

	// FUNCTION BODY
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		log.Printf("Error generating password hash: %s", err)
		w.WriteHeader(500)
		return
	}
	user, err := cfg.db.UpdateUser(id, request.Email, hash)
	if err != nil {
		log.Printf("Error updating user: %s", err)
		w.WriteHeader(500)
		return
	}

	// RESPONSE
	type responseStruct struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
	}
	writeResponse(w, 200, responseStruct{
		Email: user.Email,
		ID:    user.ID,
	})
}

func (cfg *apiConfig) postRefreshHandler(w http.ResponseWriter, r *http.Request) {
	// CHECKING AUTHENTICATION
	id, err := auth.AuthenticateRefreshToken(r, cfg.jwtSecret, cfg.db)
	if err != nil {
		writeError(w, 401, "Inavlid Token. Please log in again")
		return
	}

	// CREATE JWT TOKENS
	accessToken, err := auth.IssueAccessToken(id, cfg.jwtSecret)
	if err != nil {
		log.Printf("Error Creating Access Token: %s", err)
		w.WriteHeader(500)
		return
	}

	// RESPONSE
	type responseStruct struct {
		Token string `json:"token"`
	}
	writeResponse(w, 200, responseStruct{
		Token: accessToken,
	})
}

func (cfg *apiConfig) postRevokeHandler(w http.ResponseWriter, r *http.Request) {
	err := cfg.db.RevokeToken(strings.Split(r.Header.Get("Authorization"), " ")[1])
	if err != nil {
		log.Printf("Error Revoking Token: %s", err)
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
	}
}

func (cfg *apiConfig) validateChirp(body string, userID int) (db.Chirp, error) {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(body, " ")
	for i, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	chirp, err := cfg.db.CreateChirp(strings.Join(words, " "), userID)
	return chirp, err
}

func decodeRequest[T any](w http.ResponseWriter, r *http.Request, _ T) (T, error) {
	var request T
	var zeroVal T
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return zeroVal, err
	}
	return request, nil
}

func writeResponse[T any](w http.ResponseWriter, responseCode int, body T) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(responseCode)
	w.Write(data)
}

func writeError(w http.ResponseWriter, responseCode int, errorText string) {
	type ReturnVals struct {
		Error string `json:"error"`
	}
	data, err := json.Marshal(ReturnVals{Error: errorText})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(responseCode)
	w.Write(data)
}
