package api

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
const idLength = 8

type loginRequest struct {
	Name string `json:"name"`
}

type loginResponse struct {
	Identifier string `json:"identifier"`
}

func (rt *_router) postSession(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if len(body.Name) < 3 || len(body.Name) > 16 {
		http.Error(w, "invalid name length", http.StatusBadRequest)
		return
	}
	resp := loginResponse{Identifier: generateIdentifier(body.Name)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

/*
UserId:

	type: string
	description: Unique identifier of the user
	pattern: '^usr_[A-Za-z0-9_-]{3,60}$'
	minLength: 7      # "usr_" (4) + minimo 3 = 7
	maxLength: 64
	example: 'usr_a1B2c3'
*/
func generateIdentifier(name string) string {
	var id = "usr_"

	// Genera gli 8 caratteri casuali
	result := make([]byte, idLength)
	for i := 0; i < idLength; i++ {
		// rand.Intn() ora usa il seme impostato da init()
		randomIndex := rand.Intn(len(charSet))
		result[i] = charSet[randomIndex]
	}

	id += string(result)
	return id
}
