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
	// 1. Autenticazione (Chi manda il messaggio?)
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

	// 2. Parametri Path (In quale conversazione?)
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])

	// 3. Parsing Body (Cosa c'è scritto?)
	var req schemas.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validazione: deve esserci almeno il testo (per ora ignoriamo i media)
	if req.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	// 4. Costruzione Messaggio Completo
	// Generiamo ID e Timestamp qui
	newMessage := schemas.Message{
		ID:     generateMessageId(),
		Sender: senderId,
		Text:   req.Text,
		Time:   time.Now(),
		Status: "sent",
		Kind:   "normal",
	}

	// 5. Salvataggio nel DB
	err := rt.db.CreateMessage(convId, newMessage)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error sending message")
		// Se l'errore è "user not authorized...", potremmo dare 403, ma per ora 500 o 400
		http.Error(w, "Error sending message (are you a participant?)", http.StatusBadRequest)
		return
	}

	// 6. Risposta 204 No Content (Come da specifica)
	w.WriteHeader(http.StatusNoContent)
}

// Generatore ID casareccio per messaggi
func generateMessageId() schemas.MessageId {
	const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	const length = 10
	b := make([]byte, length)
	for i := range b {
		b[i] = charSet[rand.Intn(len(charSet))]
	}
	return schemas.MessageId("msg_" + string(b))
}
