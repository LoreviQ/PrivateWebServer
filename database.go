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
	mu     *sync.RWMutex `json:"-"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func initialiseDatabase(dbPath string) Database {
	db := Database{
		dbPath: dbPath,
		Chirps: make(map[int]Chirp),
		mu:     &sync.RWMutex{},
	}
	err := db.ensureDB()
	if err != nil {
		log.Fatal(err)
	}
	err = db.loadDB()
	if err != nil {
		log.Fatal(err)
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
	err = os.WriteFile("database/database.json", data, 0777)
	return err
}
