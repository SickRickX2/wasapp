package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) getMessages(w http.ResponseWriter, r *http.Request) {
	// prende i parametri dall'url
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])

	// autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// prende i parametri
	limit := 20
	// default
	if l := r.URL.Query().Get("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	beforeId := r.URL.Query().Get("beforeId")

	// chiama il db
	msgs, err := rt.db.GetConversationMessages(convId, limit, beforeId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving messages")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// risposta con i messaggi
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}
