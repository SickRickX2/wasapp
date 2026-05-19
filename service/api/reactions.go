package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

// PUT /conversations/{convId}/messages/{messageId}/reaction
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))

	vars := ps
	convId := schemas.ConversationId(vars.ByName("convId"))
	messageId := schemas.MessageId(vars.ByName("messageId"))
	if convId == "" || messageId == "" {
		http.Error(w, "Invalid path parameters", http.StatusBadRequest)
		return
	}

	//utenete e messaggio nella conv
	isInConversation, err := rt.db.IsUserInConversation(convId, userId)
	if err != nil || !isInConversation {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	messageInConversation, err := rt.db.IsMessageInConversation(convId, messageId)
	if err != nil || !messageInConversation {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// parsa il body
	var req struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	emojiLen := utf8.RuneCountInString(req.Emoji)
	if emojiLen < 1 || emojiLen > 4 {
		http.Error(w, "Emoji length must be between 1 and 4", http.StatusBadRequest)
		return
	}

	// aggiorna il db
	err = rt.db.ReactToMessage(messageId, userId, req.Emoji)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// recupera il messaggio aggiornato dal DB
	updatedMsg, err := rt.db.GetMessage(messageId)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedMsg); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in commentMessage")
	}
}

// DELETE /conversations/{convId}/messages/{messageId}/reaction
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))
	vars := ps
	convId := schemas.ConversationId(vars.ByName("convId"))
	messageId := schemas.MessageId(vars.ByName("messageId"))
	if convId == "" || messageId == "" {
		http.Error(w, "Invalid path parameters", http.StatusBadRequest)
		return
	}
	// messaggio e user nella conv
	isInConversation, err := rt.db.IsUserInConversation(convId, userId)
	if err != nil || !isInConversation {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	messageInConversation, err := rt.db.IsMessageInConversation(convId, messageId)
	if err != nil || !messageInConversation {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// rimuove dal db
	err = rt.db.UnreactToMessage(messageId, userId)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	updatedMsg, err := rt.db.GetMessage(messageId)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedMsg); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in uncommentMessage")
	}
}
