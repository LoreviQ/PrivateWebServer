package db

type Chirp struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

func (db *Database) CreateChirp(chirpText string) (Chirp, error) {
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
