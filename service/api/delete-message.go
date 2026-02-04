package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request) {
	// autenticazione
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

	// prende i parametri dall'url
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])
	messageId := schemas.MessageId(vars["messageId"])

	// cancella dal db
	updatedMsg, err := rt.db.DeleteMessage(convId, messageId, userId)
	if err != nil {
		// se non l'ha trovato o non è autorizzato  restituisce 404
		rt.baseLogger.WithError(err).Error("Error deleting message")
		http.Error(w, "Message not found or unauthorized", http.StatusNotFound)
		return
	}

	// tocca aggiornare la risposta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedMsg)
}
