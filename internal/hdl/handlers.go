package hdl

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/LoreviQ/PrivateWebServer/internal/auth"
	"github.com/LoreviQ/PrivateWebServer/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type ApiConfig struct {
	Port           string
	DB_Directory   string
	JWT_Secret     []byte
	FileserverHits int
	DB             db.Database
}

func (cfg *ApiConfig) HandleFlags() {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		log.Printf("Entering debug mode\n")
		cfg.DB_Directory = "./database/debugDB.json"
		os.Remove(cfg.DB_Directory)
	}
}

func (cfg *ApiConfig) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Status", "200 OK")
	w.Write([]byte("OK"))
}

func (cfg *ApiConfig) MetricsReportingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Status", "200 OK")
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %v times!</p></body></html>", cfg.FileserverHits)))
}

func (cfg *ApiConfig) MetricsResetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Status", "200 OK")
	cfg.FileserverHits = 0
}

func (cfg *ApiConfig) GetChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirps := make([]db.Chirp, len(cfg.DB.Chirps))
	for i, chirp := range cfg.DB.Chirps {
		chirps[i-1] = chirp
	}
	writeResponse(w, 200, chirps)
}

func (cfg *ApiConfig) GetChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, 400, "Invalid ID")
		return
	}
	chirp, ok := cfg.DB.Chirps[id]
	if !ok {
		writeError(w, 404, "No Chirp by that ID")
		return
	}
	writeResponse(w, 200, chirp)
}

func (cfg *ApiConfig) PostChirpHandler(w http.ResponseWriter, r *http.Request) {
	// CHECKING AUTHENTICATION
	id, err := auth.AuthenticateAccessToken(r, cfg.JWT_Secret)
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

func (cfg *ApiConfig) DeleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	// AUTHENTICATION
	userID, err := auth.AuthenticateAccessToken(r, cfg.JWT_Secret)
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
	chirp, ok := cfg.DB.Chirps[chirpID]
	if !ok {
		writeError(w, 404, "No Chirp by that ID")
		return
	}
	if chirp.UserID != userID {
		writeError(w, 403, "Not Authorised to delete this chirp")
		return
	}

	// DELETING CHIRP
	err = cfg.DB.DeleteChirp(chirpID)
	if err != nil {
		log.Printf("Error Deleting Chirp: %s", err)
		w.WriteHeader(500)
		return
	}

	// RESPONSE
	w.WriteHeader(200)
}

func (cfg *ApiConfig) PostUserHandler(w http.ResponseWriter, r *http.Request) {
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
	user, err := cfg.DB.AddUser(request.Email, hash)
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

func (cfg *ApiConfig) PostLoginHandler(w http.ResponseWriter, r *http.Request) {
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
	user, err := cfg.DB.AuthenticateUser(request.Email, []byte(request.Password))
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
	accessToken, err := auth.IssueAccessToken(user.ID, cfg.JWT_Secret)
	if err != nil {
		log.Printf("Error Creating Access Token: %s", err)
		w.WriteHeader(500)
		return
	}
	refreshToken, err := auth.IssueRefreshToken(user.ID, cfg.JWT_Secret, cfg.DB)
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

func (cfg *ApiConfig) PutUserHandler(w http.ResponseWriter, r *http.Request) {
	// CHECKING AUTHENTICATION
	id, err := auth.AuthenticateAccessToken(r, cfg.JWT_Secret)
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
	user, err := cfg.DB.UpdateUser(id, request.Email, hash)
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

func (cfg *ApiConfig) PostRefreshHandler(w http.ResponseWriter, r *http.Request) {
	// CHECKING AUTHENTICATION
	id, err := auth.AuthenticateRefreshToken(r, cfg.JWT_Secret, cfg.DB)
	if err != nil {
		writeError(w, 401, "Inavlid Token. Please log in again")
		return
	}

	// CREATE JWT TOKENS
	accessToken, err := auth.IssueAccessToken(id, cfg.JWT_Secret)
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

func (cfg *ApiConfig) PostRevokeHandler(w http.ResponseWriter, r *http.Request) {
	err := cfg.DB.RevokeToken(strings.Split(r.Header.Get("Authorization"), " ")[1])
	if err != nil {
		log.Printf("Error Revoking Token: %s", err)
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
	}
}

func (cfg *ApiConfig) validateChirp(body string, userID int) (db.Chirp, error) {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(body, " ")
	for i, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	chirp, err := cfg.DB.CreateChirp(strings.Join(words, " "), userID)
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
