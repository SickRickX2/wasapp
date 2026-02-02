package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) setUsername(w http.ResponseWriter, r *http.Request) {
	// 1. Prendi userId dal path
	vars := mux.Vars(r)
	userId := schemas.UserId(vars["userId"])

	// 2. Decodifica il body
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// 3. Validazione (Regex + Lunghezza come da api.yaml)
	if len(req.Username) < 3 || len(req.Username) > 16 {
		http.Error(w, "Username length must be between 3 and 16", http.StatusBadRequest)
		return
	}
	// Regex: solo lettere, numeri, trattini e underscore
	validName := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	if !validName.MatchString(req.Username) {
		http.Error(w, "Invalid characters in username", http.StatusBadRequest)
		return
	}

	// 4. Aggiorna nel Database
	err := rt.db.SetUserName(userId, req.Username)
	if err != nil {
		// Potrebbe fallire se il nome è già preso (UNIQUE constraint)
		rt.baseLogger.WithError(err).Error("Error updating username")
		http.Error(w, "Internal Server Error (Name taken?)", http.StatusInternalServerError)
		return
	}

	// 5. Recupera l'utente aggiornato per restituirlo
	updatedUser, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Error retrieving updated user", http.StatusInternalServerError)
		return
	}

	// 6. Rispondi con l'oggetto User aggiornato
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}
