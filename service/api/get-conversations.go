package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	requestingUserId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))
	targetUserId := schemas.UserId(ps.ByName("userId"))
	if targetUserId == "" {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}
	if requestingUserId != targetUserId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	// prende le conversazioni dal db
	conversations, err := rt.db.GetConversations(targetUserId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving conversations")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// risposta
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Conversations []schemas.Conversation `json:"conversations"`
	}{
		// Se conversations è nil (es. zero chat), restituiamo un array vuoto invece di null
		Conversations: conversations,
	}
	if response.Conversations == nil {
		response.Conversations = make([]schemas.Conversation, 0)
	}

	// 2. Facciamo l'Encode e controlliamo l'errore per accontentare il linter!
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in getConversations")
	}
}
