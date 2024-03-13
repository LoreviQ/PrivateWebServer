package hdl

import (
	"errors"
	"log"
	"net/http"

	"github.com/LoreviQ/PrivateWebServer/internal/auth"
	"github.com/LoreviQ/PrivateWebServer/internal/db"
	"golang.org/x/crypto/bcrypt"
)

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
		Email     string `json:"email"`
		ID        int    `json:"id"`
		ChirpyRed bool   `json:"is_chirpy_red"`
	}
	writeResponse(w, 201, responseStruct{
		Email:     user.Email,
		ID:        user.ID,
		ChirpyRed: user.ChirpyRed,
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
		Email     string `json:"email"`
		ID        int    `json:"id"`
		ChirpyRed bool   `json:"is_chirpy_red"`
	}
	writeResponse(w, 200, responseStruct{
		Email:     user.Email,
		ID:        user.ID,
		ChirpyRed: user.ChirpyRed,
	})
}
