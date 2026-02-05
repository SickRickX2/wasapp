package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, "Bearer "))

	// 2. Parametri Path (Sorgente)
	vars := mux.Vars(r)
	sourceConvId := schemas.ConversationId(vars["convId"])
	originalMessageId := schemas.MessageId(vars["messageId"])

	// 3. Parsing Body (Destinazione)
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

	// 4. CHECK SICUREZZA 1: L'utente è nella chat SORGENTE?
	inSource, err := rt.db.IsUserInConversation(sourceConvId, userId)
	if err != nil {
		http.Error(w, "Error checking source permissions", http.StatusInternalServerError)
		return
	}
	if !inSource {
		http.Error(w, "Forbidden: You cannot forward messages from a chat you are not in", http.StatusForbidden)
		return
	}

	// 5. CHECK SICUREZZA 2: L'utente è nella chat DESTINAZIONE?
	inDest, err := rt.db.IsUserInConversation(req.DestinationConvId, userId)
	if err != nil {
		http.Error(w, "Error checking destination permissions", http.StatusInternalServerError)
		return
	}
	if !inDest {
		http.Error(w, "Forbidden: You cannot forward messages to a chat you are not in", http.StatusForbidden)
		return
	}

	// 6. Recupera il messaggio originale
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

	// 7. Crea il NUOVO messaggio (Copia)
	forwardedMsg := schemas.Message{
		ID:      generateMessageId(), // Nuovo ID
		Sender:  userId,              // Il mittente sei TU (che inoltri), non l'autore originale
		Text:    originalMsg.Text,    // Copia testo
		MediaId: originalMsg.MediaId, // Copia media
		Time:    time.Now(),
		Status:  "sent",
		Kind:    "forwarded", // <--- Importante!
	}

	// 8. Salva nel DB (Destinazione)
	// Riutilizziamo la funzione CreateMessage che abbiamo già!
	err = rt.db.CreateMessage(req.DestinationConvId, forwardedMsg)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error saving forwarded message")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 9. Risposta (Restituisce il nuovo messaggio creato)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(forwardedMsg)
}
