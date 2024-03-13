package hdl

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/LoreviQ/PrivateWebServer/internal/auth"
	"github.com/LoreviQ/PrivateWebServer/internal/db"
)

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
