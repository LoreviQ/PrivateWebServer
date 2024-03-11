package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type Database struct {
	dbPath string         `json:"-"`
	Chirps map[int]chirps `json:"chirps"`
	mu     *sync.RWMutex  `json:"-"`
}

type chirps struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func initialiseDatabase(dbPath string) Database {
	db := Database{dbPath: dbPath}
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

func (db *Database) ensureDB() error {
	_, err := os.ReadFile(db.dbPath)
	if errors.Is(err, os.ErrNotExist) {
		err = db.writeDB()
	}
	return err
}

func (db *Database) loadDB() error {
	return nil
}

func (db *Database) writeDB() error {
	data, err := json.Marshal(db)
	if err != nil {
		return err
	}
	err = os.WriteFile("database/database.json", data, 0777)
	return err
}
