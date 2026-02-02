package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
const idLength = 8

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	// 1. Decodifica JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// 2. Validazione
	if len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, "Username must be between 3 and 16 chars", http.StatusBadRequest)
		return
	}

	var identifier schemas.UserId

	// 3. Cerca se l'utente esiste già
	user, err := rt.db.FindUserByName(req.Name)
	if err == nil {
		// Trovato!
		identifier = user.ID
	} else {
		// Non trovato (o errore). Assumiamo che se c'è errore, l'utente non c'è.
		// Iniziamo il ciclo di creazione con retry (per evitare ID duplicati)
		const maxRetries = 3
		var creationSuccess bool

		for i := 0; i < maxRetries; i++ {
			// Genera un ID casuale tipo 'usr_XyZ123'
			newID := generateUserId()

			newUser := schemas.User{
				ID:        newID,
				Name:      req.Name,
				CreatedAt: time.Now(),
			}

			// Prova a inserire nel DB
			err = rt.db.CreateUser(newUser)
			if err == nil {
				// Successo!
				identifier = newID
				creationSuccess = true
				break
			}

			// Se fallisce, probabilmente l'ID esiste già (molto raro con 8 char, ma possibile).
			// Il ciclo continua e ne prova un altro.
		}

		if !creationSuccess {
			// Se dopo 3 tentativi fallisce ancora, è un problema serio del DB
			rt.baseLogger.WithError(err).Error("Failed to create user after retries")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// 4. Rispondi con successo (201 Created)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Identifier schemas.UserId `json:"identifier"`
	}{Identifier: identifier})
}

// generateUserId crea un ID che rispetta il pattern '^usr_[A-Za-z0-9_-]{3,60}$'
func generateUserId() schemas.UserId {

	idPrefix := "usr_"

	result := make([]byte, idLength)
	for i := 0; i < idLength; i++ {

		randomIndex := rand.Intn(len(charSet))
		result[i] = charSet[randomIndex]
	}

	return schemas.UserId(idPrefix + string(result))
}
