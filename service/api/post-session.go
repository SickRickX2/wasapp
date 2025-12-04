package api

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"

	"github.com/SickRickX2/wasapp/service/database"
)

const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
const idLength = 8

type loginRequest struct {
	Name string `json:"name"`
}

type loginResponse struct {
	Identifier string `json:"identifier"`
}

func (rt *_router) postSession(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if len(body.Name) < 3 || len(body.Name) > 16 {
		http.Error(w, "invalid name length", http.StatusBadRequest)
		return
	}
	var identifier string
	id, err := rt.db.FindUserByName(body.Name)
	if err == nil {
		identifier = id
	} else if errors.Is(err, database.ErrUserNotFound) {
		const maxRetries = 3
		var creationErr error
		for i := 0; i < maxRetries; i++ {
			identifier = generateIdentifier() // Genera ID casuale (senza 'name')

			creationErr = rt.db.CreateUser(identifier, body.Name)

			// Se l'inserimento ha successo, usciamo dal ciclo.
			if creationErr == nil {
				break
			}

			// Se l'errore NON è una violazione di chiave primaria, è un errore critico (es. connessione)
			if !errors.Is(err, database.ErrDuplicateKey) {
				http.Error(w, "error creating user: critical DB error", http.StatusInternalServerError)
				return
			}
			// Se è una violazione di chiave, il loop continua per un nuovo tentativo.
		}

		if creationErr != nil && errors.Is(creationErr, database.ErrDuplicateKey) {
			http.Error(w, "error creating user: failed to find unique ID after retries", http.StatusInternalServerError)
			return
		}
	} else {
		// Errore generico durante la ricerca (es. connessione)
		http.Error(w, "Error querying user", http.StatusInternalServerError)
		return
	}

	resp := loginResponse{Identifier: identifier}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
	/*
		resp := loginResponse{Identifier: generateIdentifier(body.Name)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	*/
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
func generateIdentifier() string {
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
