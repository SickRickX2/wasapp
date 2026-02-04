package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticazione (Chi crea il gruppo?)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	creatorId := schemas.UserId(parts[1])

	// 2. Parsing Body
	var req struct {
		GroupName    string           `json:"groupName"`
		Participants []schemas.UserId `json:"participants"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 3. Validazione
	if len(req.GroupName) < 1 || len(req.GroupName) > 30 {
		http.Error(w, "Group name must be between 1 and 30 chars", http.StatusBadRequest)
		return
	}
	// Possiamo anche controllare che la lista non sia vuota, ma non è obbligatorio

	// 4. Chiama DB
	group, err := rt.db.CreateGroup(creatorId, req.GroupName, req.Participants)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error creating group")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Risposta
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}
