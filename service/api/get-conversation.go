package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

// GET /conversations/{convId}
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))
	convId := schemas.ConversationId(ps.ByName("convId"))

	if convId == "" {
		http.Error(w, "Invalid convId", http.StatusBadRequest)
		return
	}

	isInConv, err := rt.db.IsUserInConversation(convId, userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error checking conversation membership in getConversation")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isInConv {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	conversation, err := rt.db.GetConversation(convId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving conversation in getConversation")
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conversation); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in getConversation")
	}
}
