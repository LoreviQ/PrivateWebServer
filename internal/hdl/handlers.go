package hdl

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/LoreviQ/PrivateWebServer/internal/db"
)

type ApiConfig struct {
	Port           string
	DB_Directory   string
	JWT_Secret     []byte
	FileserverHits int
	DB             db.Database
}

func (cfg *ApiConfig) HandleFlags() {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		log.Printf("Entering debug mode\n")
		cfg.DB_Directory = "./database/debugDB.json"
		os.Remove(cfg.DB_Directory)
	}
}

func decodeRequest[T any](w http.ResponseWriter, r *http.Request, _ T) (T, error) {
	var request T
	var zeroVal T
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return zeroVal, err
	}
	return request, nil
}

func writeResponse[T any](w http.ResponseWriter, responseCode int, body T) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(responseCode)
	w.Write(data)
}

func writeError(w http.ResponseWriter, responseCode int, errorText string) {
	type ReturnVals struct {
		Error string `json:"error"`
	}
	data, err := json.Marshal(ReturnVals{Error: errorText})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(responseCode)
	w.Write(data)
}
