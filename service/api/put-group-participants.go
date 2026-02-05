package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parametri Path (Solo ID Conversazione)
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])

	// 3. Parsing Body (Lista utenti)
	var req struct {
		UserIds []schemas.UserId `json:"userIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validazione minima: lista non vuota
	if len(req.UserIds) == 0 {
		http.Error(w, "User list cannot be empty", http.StatusBadRequest)
		return
	}

	// 4. Chiama il DB
	err := rt.db.AddGroupMembers(convId, req.UserIds)
	if err != nil {
		if err.Error() == "cannot add members to a private conversation" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "no rows") { // Check veloce se conv non trovata
			http.Error(w, "Conversation not found", http.StatusNotFound)
		} else {
			rt.baseLogger.WithError(err).Error("Error adding members to group")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// 5. Successo
	w.WriteHeader(http.StatusOK) // O 200 OK come da tuo YAML
	// Opzionale: restituire il gruppo aggiornato come promette lo YAML,
	// ma per ora va bene anche solo status OK.
}
