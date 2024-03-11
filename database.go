package main

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type Database struct {
	dbPath string
	Chirps map[int]chirps
	mu     *sync.RWMutex
}

type chirps struct {
	ID   int
	Body string
}

func initialiseDatabase(dbPath string) Database {
	db := Database{dbPath: dbPath}
	err := db.ensureDB()
	fmt.Print(err)
	/*
		if err != nil {
			db.writeDB()
		}
		db.loadDB()
	*/
	return db
}

func (db *Database) ensureDB() error {
	_, err := os.ReadFile(db.dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

/*
func (db *Database) loadDB() error

func (db *Database) writeDB() error
*/
