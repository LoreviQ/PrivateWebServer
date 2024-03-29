package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LoreviQ/PrivateWebServer/internal/db"
	"github.com/golang-jwt/jwt/v5"
)

func IssueAccessToken(userID int, secret []byte) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Subject:   fmt.Sprint(userID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	return signedToken, err
}

func IssueRefreshToken(userID int, secret []byte, db db.Database) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1440)),
		Subject:   fmt.Sprint(userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	err = db.AddToken(signedToken)
	if err != nil {
		return "", err
	}

	return signedToken, err
}

func AuthenticateAccessToken(r *http.Request, secret []byte) (int, error) {
	tokenString := strings.Split(r.Header.Get("Authorization"), " ")[1]
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("token invalid")
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil || issuer != "chirpy-access" {
		return 0, errors.New("issuer invalid")
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return 0, err
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}

	return idInt, nil
}

func AuthenticateRefreshToken(r *http.Request, secret []byte, db db.Database) (int, error) {
	tokenString := strings.Split(r.Header.Get("Authorization"), " ")[1]
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("token invalid")
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil || issuer != "chirpy-refresh" {
		return 0, errors.New("issuer invalid")
	}

	if !db.Tokens[tokenString].Valid {
		return 0, errors.New("token has been revoked")
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return 0, err
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}

	return idInt, nil
}

func AuthenticateAPI(r *http.Request, expected string) error {
	apiKey := strings.Split(r.Header.Get("Authorization"), " ")
	if apiKey[0] != "ApiKey" {
		return errors.New("wrong auth type")
	}
	if apiKey[1] != expected {
		return errors.New("invalid api key")
	}
	return nil
}
