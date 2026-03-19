package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

var validUsernameRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func (rt *_router) setUsername(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenUserId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))

	// path params
	vars := ps
	userId := schemas.UserId(vars.ByName("userId"))

	// 3controlla sicurezza
	if tokenUserId != userId {
		http.Error(w, "Forbidden: You can only change your own username", http.StatusForbidden)
		return
	}

	// legge il body
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// validazioni usando la regex globale
	if len(req.Username) < 3 || len(req.Username) > 16 {
		http.Error(w, "Username length must be between 3 and 16", http.StatusBadRequest)
		return
	}
	if !validUsernameRegex.MatchString(req.Username) {
		http.Error(w, "Invalid characters in username", http.StatusBadRequest)
		return
	}

	// aggiorna nel db
	err := rt.db.SetUserName(userId, req.Username)
	if err != nil {
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
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in setUsername")
	}
}
