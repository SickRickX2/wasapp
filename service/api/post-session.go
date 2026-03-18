package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
const idLength = 8

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, "Username must be between 3 and 16 chars", http.StatusBadRequest)
		return
	}

	var identifier schemas.UserId

	//  cerca se esiste
	user, err := rt.db.FindUserByName(req.Name)
	if err == nil {

		identifier = user.ID
	} else {
		// se non esiste, lo crea
		const maxRetries = 3
		var creationSuccess bool

		for i := 0; i < maxRetries; i++ {
			// genera un suo id
			newID := generateUserId()

			newUser := schemas.User{
				ID:        newID,
				Name:      req.Name,
				CreatedAt: time.Now(),
			}

			// inserisce nel db
			err = rt.db.CreateUser(newUser)
			if err == nil {

				identifier = newID
				creationSuccess = true
				break
			}

		}

		if !creationSuccess {
			// se dopo i tentativi non riesce allora errore
			rt.baseLogger.WithError(err).Error("Failed to create user after retries")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// risponde con l'id
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
