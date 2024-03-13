package hdl

import (
	"log"
	"net/http"
)

func (cfg *ApiConfig) PostPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	// REQUEST
	type requestStruct struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}
	request, err := decodeRequest(w, r, requestStruct{})
	if err != nil {
		log.Printf("Error decoding Request: %s", err)
		w.WriteHeader(500)
		return
	}

	// UPDATING USER TO CHIRPY RED
	if request.Event != "user.upgraded" {
		w.WriteHeader(200)
		return
	}

	err = cfg.DB.AddChirpyRed(request.Data.UserID)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	// RESPONSE
	w.WriteHeader(200)
}
