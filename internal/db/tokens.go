package db

import "time"

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
