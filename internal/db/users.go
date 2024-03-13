package db

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrTakenEmail = errors.New("email already taken")
var ErrInvalidEmail = errors.New("invalid email address")
var ErrIncorrectPassword = errors.New("inocrrect password")
var ErrInvalidUserID = errors.New("invalid user ID")

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"hash"`
	ChirpyRed    bool   `json:"is_chirpy_red"`
}

func (db *Database) AddUser(email string, hash []byte) (User, error) {
	db.mu.Lock()
	for _, user := range db.Users {
		if user.Email == email {
			db.mu.Unlock()
			return User{}, ErrTakenEmail
		}
	}
	id := len(db.Users) + 1
	db.Users[id] = User{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
		ChirpyRed:    false,
	}
	db.mu.Unlock()
	err := db.writeDB()
	if err != nil {
		return User{}, err
	}
	return db.Users[id], err
}

func (db *Database) UpdateUser(id int, email string, hash []byte) (User, error) {
	db.mu.Lock()
	db.Users[id] = User{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
		ChirpyRed:    db.Users[id].ChirpyRed,
	}
	db.mu.Unlock()
	err := db.writeDB()
	if err != nil {
		return User{}, err
	}
	return db.Users[id], err
}

func (db *Database) AuthenticateUser(email string, password []byte) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	id := 0
	for i, user := range db.Users {
		if user.Email == email {
			id = i
		}
	}
	if id == 0 {
		return User{}, ErrInvalidEmail
	}
	err := bcrypt.CompareHashAndPassword(db.Users[id].PasswordHash, password)
	if err != nil {
		return User{}, ErrIncorrectPassword
	}

	return db.Users[id], nil
}

func (db *Database) AddChirpyRed(userID int) error {
	db.mu.Lock()
	_, ok := db.Users[userID]
	if !ok {
		return ErrInvalidUserID
	}
	db.Users[userID] = User{
		ID:           userID,
		Email:        db.Users[userID].Email,
		PasswordHash: db.Users[userID].PasswordHash,
		ChirpyRed:    true,
	}
	db.mu.Unlock()
	err := db.writeDB()
	return err
}
