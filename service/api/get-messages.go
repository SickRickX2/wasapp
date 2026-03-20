package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	convId := schemas.ConversationId(ps.ByName("convId"))

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, "Bearer "))

	// controlla se l'utente è nella conversazione
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

	// prende i parametri di paginazione
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	beforeId := r.URL.Query().Get("beforeId")

	// chiama il db per i messaggi
	msgs, err := rt.db.GetConversationMessages(convId, limit, beforeId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving messages")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for i := range msgs {
		if msgs[i].MediaId == "" {
			continue
		}

		url, err := rt.db.GetMediaUrl(msgs[i].MediaId)
		if err != nil {
			continue
		}

		msgs[i].Media = &schemas.Media{URL: url}
	}

	// risposta
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(msgs); err != nil {
		rt.baseLogger.WithError(err).Error("Error encoding response")
		return
	}
}
