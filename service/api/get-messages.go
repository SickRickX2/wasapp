package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings" // <--- Aggiunto per gestire il Bearer Token

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) getMessages(w http.ResponseWriter, r *http.Request) {
	// 1. Prende i parametri dall'url
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])

	// 2. Autenticazione (Estraiamo lo UserId)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, "Bearer "))

	// 3. NUOVO: Controllo Sicurezza (Sei parte della chat?)
	// Usiamo la funzione che hai già creato in check-membership.go
	isInConv, err := rt.db.IsUserInConversation(convId, userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error checking conversation membership")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isInConv {
		http.Error(w, "Forbidden: You are not a participant of this conversation", http.StatusForbidden)
		return
	}

	// 4. Prende i parametri di paginazione
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	beforeId := r.URL.Query().Get("beforeId")

	// 5. Chiama il db per i messaggi
	msgs, err := rt.db.GetConversationMessages(convId, limit, beforeId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving messages")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 6. Risposta
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(msgs); err != nil {
		rt.baseLogger.WithError(err).Error("Error encoding response")
		return
	}
}
