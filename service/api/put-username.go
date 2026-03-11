package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setUsername(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// prende l'utente dal path
	vars := ps
	userId := schemas.UserId(vars.ByName("userId"))

	// legge il body
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// validazioni
	if len(req.Username) < 3 || len(req.Username) > 16 {
		http.Error(w, "Username length must be between 3 and 16", http.StatusBadRequest)
		return
	}
	// regex
	validName := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	if !validName.MatchString(req.Username) {
		http.Error(w, "Invalid characters in username", http.StatusBadRequest)
		return
	}

	// aggiorna nel db
	err := rt.db.SetUserName(userId, req.Username)
	if err != nil {
		// se il nome è già preso allora erore
		rt.baseLogger.WithError(err).Error("Error updating username")
		http.Error(w, "Internal Server Error (Name taken?)", http.StatusInternalServerError)
		return
	}

	// prendi l'utente
	updatedUser, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Error retrieving updated user", http.StatusInternalServerError)
		return
	}

	// rispondi con l'utente aggiornato
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}
