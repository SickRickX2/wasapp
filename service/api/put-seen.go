package api

import (
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) markAsSeen(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, "Bearer "))

	// 2. Parametri Path
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])
	messageId := schemas.MessageId(vars["messageId"])

	// 3. Controllo Sicurezza: Sono nella chat?
	// (Opzionale ma consigliato, anche se MarkAsSeen filtra già per senderId != me)
	inConv, err := rt.db.IsUserInConversation(convId, userId)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !inConv {
		http.Error(w, "Forbidden: You are not in this conversation", http.StatusForbidden)
		return
	}

	// 4. Chiama DB
	err = rt.db.MarkAsSeen(convId, messageId, userId)
	if err != nil {
		if err.Error() == "message not found" {
			http.Error(w, "Message not found", http.StatusNotFound)
		} else {
			rt.baseLogger.WithError(err).Error("Error marking message as seen")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// 5. Risposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Rispondiamo con un JSON semplice come da esempio YAML (o empty se preferisci)
	w.Write([]byte(`{"success": true}`))
}
