package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request) {
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
	senderId := schemas.UserId(parts[1])

	// prende la conversazione
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])

	// prende ilmessaggio dal body
	var req struct {
		Text    string `json:"text"`
		MediaId string `json:"mediaId"` // Campo opzionale
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// deve esserci almeno uno tra testo o media TODO: media non implementato
	if req.Text == "" && req.MediaId == "" {
		http.Error(w, "Message must contain either text or media", http.StatusBadRequest)
		return
	}

	// creazione del messaggio
	newMessage := schemas.Message{
		ID:      generateMessageId(),
		Sender:  senderId,
		Text:    req.Text,
		MediaId: req.MediaId,
		Time:    time.Now(),
		Status:  "sent",
		Kind:    "normal",
	}

	// salva nel db
	err := rt.db.CreateMessage(convId, newMessage)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error sending message")
		http.Error(w, "Error sending message (are you a participant?)", http.StatusBadRequest)
		return
	}

	// risposta di successo
	w.WriteHeader(http.StatusNoContent)
}

// genera un id casuale per il messaggio
func generateMessageId() schemas.MessageId {
	const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	const length = 10
	b := make([]byte, length)
	for i := range b {
		b[i] = charSet[rand.Intn(len(charSet))]
	}
	return schemas.MessageId("msg_" + string(b))
}
