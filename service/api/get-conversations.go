package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (rt *_router) getConversations(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, "Bearer "))

	// 2. Chiama DB
	conversations, err := rt.db.GetConversations(userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving conversations")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 3. Rispondi
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}
