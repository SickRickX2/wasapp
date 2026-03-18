package api

import (
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) markAsSeen(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// authentication
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))
	// path params
	vars := ps
	convId := schemas.ConversationId(vars.ByName("convId"))
	messageId := schemas.MessageId(vars.ByName("messageId"))

	// controllo sicurezza
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
