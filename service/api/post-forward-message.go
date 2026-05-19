package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, "Bearer "))

	vars := ps
	sourceConvId := schemas.ConversationId(vars.ByName("convId"))
	originalMessageId := schemas.MessageId(vars.ByName("messageId"))

	// parsa il body
	var req struct {
		DestinationConvId schemas.ConversationId `json:"destinationConversationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.DestinationConvId == "" {
		http.Error(w, "destinationConversationId is required", http.StatusBadRequest)
		return
	}

	// guarda se l'utente è nella chat in
	inSource, err := rt.db.IsUserInConversation(sourceConvId, userId)
	if err != nil {
		http.Error(w, "Error checking source permissions", http.StatusInternalServerError)
		return
	}
	if !inSource {
		http.Error(w, "Forbidden: You cannot forward messages from a chat you are not in", http.StatusForbidden)
		return
	}

	// guarda se l'utente è nella chat out
	inDest, err := rt.db.IsUserInConversation(req.DestinationConvId, userId)
	if err != nil {
		http.Error(w, "Error checking destination permissions", http.StatusInternalServerError)
		return
	}
	if !inDest {
		http.Error(w, "Forbidden: You cannot forward messages to a chat you are not in", http.StatusForbidden)
		return
	}

	// orende il messaggio og
	originalMsg, err := rt.db.GetMessage(originalMessageId)
	if err != nil {
		if err.Error() == "message not found" {
			http.Error(w, "Message not found", http.StatusNotFound)
		} else {
			rt.baseLogger.WithError(err).Error("Error fetching message")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// crea il nuovo messaggio da inoltrare
	forwardedMsg := schemas.Message{
		ID:      generateMessageId(),
		Sender:  userId,
		Text:    originalMsg.Text,
		MediaId: originalMsg.MediaId,
		Time:    time.Now(),
		Status:  "sent",
		Kind:    "forwarded",
	}

	// salva il messaggio inoltrato nella chat di destinazione
	err = rt.db.CreateMessage(req.DestinationConvId, forwardedMsg)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error saving forwarded message")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// risposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(forwardedMsg); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in forwardMessage")
	}
}
