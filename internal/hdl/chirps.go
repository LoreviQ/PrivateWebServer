package hdl

import (
	"log"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/LoreviQ/PrivateWebServer/internal/auth"
	"github.com/LoreviQ/PrivateWebServer/internal/db"
)

func (cfg *ApiConfig) GetChirpHandler(w http.ResponseWriter, r *http.Request) {
	// QUERY PARAMETERS
	authorID := r.URL.Query().Get("author_id")
	sortParam := r.URL.Query().Get("sort")

	// GET SLICE OF CHIRPS
	chirps := make([]db.Chirp, 0, len(cfg.DB.Chirps))
	if authorID == "" {
		for _, chirp := range cfg.DB.Chirps {
			chirps = append(chirps, chirp)
		}
	} else {
		id, err := strconv.Atoi(authorID)
		if err != nil {
			writeError(w, 400, "Invalid ID")
			return
		}
		for _, chirp := range cfg.DB.Chirps {
			if chirp.UserID == id {
				chirps = append(chirps, chirp)
			}
		}
	}

	//SORT CHIRPS
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})
	if sortParam == "desc" {
		slices.Reverse(chirps)
	}

	//RESPONSE
	writeResponse(w, 200, chirps)
}

func (cfg *ApiConfig) GetChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, 400, "Invalid ID")
		return
	}
	chirp, ok := cfg.DB.Chirps[id]
	if !ok {
		writeError(w, 404, "No Chirp by that ID")
		return
	}
	writeResponse(w, 200, chirp)
}

func (cfg *ApiConfig) PostChirpHandler(w http.ResponseWriter, r *http.Request) {
	// CHECKING AUTHENTICATION
	id, err := auth.AuthenticateAccessToken(r, cfg.JWT_Secret)
	if err != nil {
		writeError(w, 401, "Inavlid Token. Please log in again")
		return
	}

	// REQUEST
	type requestStruct struct {
		Body string `json:"body"`
	}
	request, err := decodeRequest(w, r, requestStruct{})
	if err != nil {
		return
	}

	// POST CHIRP
	w.Header().Set("Content-Type", "application/json")
	if len(request.Body) <= 140 {
		chirp, err := cfg.validateChirp(request.Body, id)
		if err != nil {
			log.Printf("Error validating chirp: %s", err)
			w.WriteHeader(500)
			return
		}

		// RESPONSE
		writeResponse(w, 201, chirp)
	} else {
		writeError(w, 400, "Chirp is too long")
	}
}

func (cfg *ApiConfig) DeleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	// AUTHENTICATION
	userID, err := auth.AuthenticateAccessToken(r, cfg.JWT_Secret)
	if err != nil {
		writeError(w, 401, "Inavlid Token. Please log in again")
		return
	}

	// AUTHORIZATION
	chirpID, err := strconv.Atoi(r.PathValue("chirpID"))
	if err != nil {
		writeError(w, 400, "Invalid ID")
		return
	}
	chirp, ok := cfg.DB.Chirps[chirpID]
	if !ok {
		writeError(w, 404, "No Chirp by that ID")
		return
	}
	if chirp.UserID != userID {
		writeError(w, 403, "Not Authorised to delete this chirp")
		return
	}

	// DELETE CHIRP
	err = cfg.DB.DeleteChirp(chirpID)
	if err != nil {
		log.Printf("Error Deleting Chirp: %s", err)
		w.WriteHeader(500)
		return
	}

	// RESPONSE
	w.WriteHeader(200)
}

func (cfg *ApiConfig) validateChirp(body string, userID int) (db.Chirp, error) {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(body, " ")
	for i, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	chirp, err := cfg.DB.CreateChirp(strings.Join(words, " "), userID)
	return chirp, err
}
