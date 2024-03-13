package db

import (
	"errors"
	"time"
)

type Token struct {
	Token          string    `json:"token"`
	Valid          bool      `json:"valid"`
	RevocationTime time.Time `json:"revocationTime"`
}

func (db *Database) AddToken(token string) error {
	db.mu.Lock()
	db.Tokens[token] = Token{
		Token: token,
		Valid: true,
	}
	db.mu.Unlock()
	err := db.writeDB()
	return err
}

func (db *Database) RevokeToken(token string) error {
	db.mu.Lock()
	_, ok := db.Tokens[token]
	if !ok {
		return errors.New("token does not exist")
	}
	if !db.Tokens[token].Valid {
		return errors.New("token already revoked")
	}
	db.Tokens[token] = Token{
		Token:          token,
		Valid:          false,
		RevocationTime: time.Now(),
	}
	db.mu.Unlock()
	err := db.writeDB()
	return err
}
