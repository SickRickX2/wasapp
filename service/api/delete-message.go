package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	appDB "github.com/SickRickX2/wasapp/service/database"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	vars := ps
	convId := schemas.ConversationId(vars.ByName("convId"))
	messageId := schemas.MessageId(vars.ByName("messageId"))

	// cancella dal db
	updatedMsg, err := rt.db.DeleteMessage(convId, messageId, userId)
	if err != nil {
		switch {
		case errors.Is(err, appDB.ErrMessageNotFound):
			http.Error(w, "Message not found", http.StatusNotFound)
		case errors.Is(err, appDB.ErrForbidden):
			http.Error(w, "Bad request", http.StatusBadRequest)
		default:
			rt.baseLogger.WithError(err).Error("Error deleting message")
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		return
	}

	// risposta 200 con messaggio aggiornato (come da YAML)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedMsg); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in deleteMessage")
	}
}
