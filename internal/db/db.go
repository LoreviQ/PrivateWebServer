package db

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type Database struct {
	dbPath string           `json:"-"`
	Chirps map[int]Chirp    `json:"chirps"`
	Users  map[int]User     `json:"users"`
	Tokens map[string]Token `json:"tokens"`
	mu     *sync.RWMutex    `json:"-"`
}

func InitialiseDatabase(dbPath string) Database {
	db := Database{
		dbPath: dbPath,
		Chirps: make(map[int]Chirp),
		Users:  make(map[int]User),
		Tokens: make(map[string]Token),
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
