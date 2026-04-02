package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	vars := ps
	convId := schemas.ConversationId(vars.ByName("convId"))
	if convId == "" {
		http.Error(w, "Invalid convId", http.StatusBadRequest)
		return
	}

	// prende il messaggio dal body
	var req struct {
		Text      string             `json:"text"`
		Media     *schemas.Media     `json:"media,omitempty"`
		MediaId   string             `json:"mediaId,omitempty"`
		ReplyToId *schemas.MessageId `json:"replyToId,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mediaId := req.MediaId
	if req.Media != nil {
		if req.Media.URL == "" {
			http.Error(w, "media.url is required", http.StatusBadRequest)
			return
		}
		parsedMediaID, err := mediaIDFromURL(req.Media.URL)
		if err != nil {
			http.Error(w, "Invalid media.url", http.StatusBadRequest)
			return
		}
		mediaId = parsedMediaID
	}

	// deve esserci almeno uno tra testo o media
	if req.Text == "" && mediaId == "" {
		http.Error(w, "Message must contain either text or media", http.StatusBadRequest)
		return
	}

	if _, err := rt.db.GetConversation(convId); err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// Determine conversation type to set initial status
	conv, err := rt.db.GetConversation(convId)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	var initialStatus string
	switch conv.(type) {
	case schemas.PrivateConversation:
		// In private conversations, message can be delivered immediately
		initialStatus = "delivered"
	case schemas.Group:
		// In groups, message also starts as delivered (no real delivery step in local app)
		initialStatus = "delivered"
	default:
		initialStatus = "delivered"
	}

	isInConversation, err := rt.db.IsUserInConversation(convId, senderId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error checking conversation membership")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isInConversation {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if mediaId != "" {
		if _, err := rt.db.GetMediaUrl(mediaId); err != nil {
			http.Error(w, "Media not found", http.StatusNotFound)
			return
		}
	}

	if req.ReplyToId != nil {
		if _, err := rt.db.GetMessage(*req.ReplyToId); err != nil {
			http.Error(w, "Reply target message not found", http.StatusNotFound)
			return
		}
	}

	// creazione del messaggio
	newMessage := schemas.Message{
		ID:        generateMessageId(),
		Sender:    senderId,
		Text:      req.Text,
		MediaId:   mediaId,
		ReplyToId: req.ReplyToId,
		Time:      time.Now(),
		Status:    initialStatus,
		Kind:      "normal",
	}

	// salva nel db
	err = rt.db.CreateMessage(convId, newMessage)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error sending message")
		http.Error(w, "Bad request", http.StatusBadRequest)
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

func mediaIDFromURL(url string) (string, error) {
	if !strings.HasPrefix(url, "/images/") {
		return "", http.ErrMissingFile
	}
	filename := strings.TrimPrefix(url, "/images/")
	if filename == "" {
		return "", http.ErrMissingFile
	}
	ext := filepath.Ext(filename)
	mediaID := strings.TrimSuffix(filename, ext)
	if !strings.HasPrefix(mediaID, "med_") || len(mediaID) <= len("med_") {
		return "", http.ErrMissingFile
	}
	return mediaID, nil
}
