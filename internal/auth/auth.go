package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LoreviQ/PrivateWebServer/internal/db"
	"github.com/golang-jwt/jwt/v5"
)

func AuthenticateAccessToken(r *http.Request, secret []byte) (int, error) {
	tokenString := strings.Split(r.Header.Get("Authorization"), " ")[1]
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		return 0, err
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
