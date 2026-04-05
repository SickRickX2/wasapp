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

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

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
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)

	// validazioni usando la regex globale
	if len(req.Username) < 3 || len(req.Username) > 16 {
		writeJSONError(w, http.StatusBadRequest, "Username length must be between 3 and 16")
		return
	}
	if !validUsernameRegex.MatchString(req.Username) {
		writeJSONError(w, http.StatusBadRequest, "Invalid characters in username")
		return
	}

	// vincolo unicità: il nome non deve essere già usato da un altro utente
	usedByOthers, err := rt.db.CountUsersByNameExcludingUser(req.Username, userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error checking username uniqueness")
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if usedByOthers > 0 {
		writeJSONError(w, http.StatusConflict, "This username is already taken by another user.")
		return
	}

	// aggiorna nel db
	err = rt.db.SetUserName(userId, req.Username)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			writeJSONError(w, http.StatusConflict, "This username is already taken by another user.")
			return
		}
		rt.baseLogger.WithError(err).Error("Error updating username")
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// prendi l'utente
	updatedUser, err := rt.db.GetUserById(userId)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Error retrieving updated user")
		return
	}

	// rispondi con l'utente aggiornato
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in setUsername")
	}
}
