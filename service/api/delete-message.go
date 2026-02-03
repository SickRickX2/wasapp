package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticazione (Chi vuole cancellare?)
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
	userId := schemas.UserId(parts[1])

	// 2. Parametri Path
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])
	messageId := schemas.MessageId(vars["messageId"])

	// 3. Chiama DB
	updatedMsg, err := rt.db.DeleteMessage(convId, messageId, userId)
	if err != nil {
		// Se non l'ha trovato o non è autorizzato, restituiamo 404 (per sicurezza non distinguiamo troppo)
		rt.baseLogger.WithError(err).Error("Error deleting message")
		http.Error(w, "Message not found or unauthorized", http.StatusNotFound)
		return
	}

	// 4. Risposta (Il messaggio aggiornato con status='deleted')
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedMsg)
}
