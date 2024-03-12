package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type Database struct {
	dbPath string        `json:"-"`
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
	mu     *sync.RWMutex `json:"-"`
}

type Chirp struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"hash"`
}

func initialiseDatabase(dbPath string) Database {
	db := Database{
		dbPath: dbPath,
		Chirps: make(map[int]Chirp),
		Users:  make(map[int]User),
		mu:     &sync.RWMutex{},
	}
	err := db.ensureDB()
	if err != nil {
		log.Panic(err)
	}
	err = db.loadDB()
	if err != nil {
		log.Panic(err)
	}
	return db
}

func (db *Database) createChirp(chirpText string) (Chirp, error) {
	db.mu.Lock()
	id := len(db.Chirps) + 1
	db.Chirps[id] = Chirp{
		ID:   id,
		Body: chirpText,
	}
	db.mu.Unlock()
	err := db.writeDB()
	if err != nil {
		var zeroVal Chirp
		return zeroVal, err
	}
	return db.Chirps[id], err
}

func (db *Database) addUser(email string, hash []byte) (User, error) {
	db.mu.Lock()
	id := len(db.Users) + 1
	db.Users[id] = User{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
	}
	db.mu.Unlock()
	err := db.writeDB()
	if err != nil {
		var zeroVal User
		return zeroVal, err
	}
	return db.Users[id], err
}

func (db *Database) ensureDB() error {
	db.mu.RLock()
	_, err := os.ReadFile(db.dbPath)

	db.mu.RUnlock()
	if errors.Is(err, os.ErrNotExist) {
		err = db.writeDB()
	}
	return err
}

func (db *Database) loadDB() error {
	db.mu.RLock()
	data, err := os.ReadFile(db.dbPath)
	db.mu.RUnlock()
	if err != nil {
		return err
	}
	db.mu.Lock()
	err = json.Unmarshal(data, &db)
	db.mu.Unlock()
	return err
}

func (db *Database) writeDB() error {
	db.mu.RLock()
	data, err := json.Marshal(db)
	db.mu.RUnlock()
	if err != nil {
		return err
	}
	err = os.WriteFile(db.dbPath, data, 0777)
	return err
}
