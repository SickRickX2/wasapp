package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request) {
	// 1. Estrai l'utente dall'Header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized: missing header", http.StatusUnauthorized)
		return
	}
	// L'header è tipo "Bearer usr_123". Splittiamo per prendere solo l'ID.
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Unauthorized: invalid token format", http.StatusUnauthorized)
		return
	}
	myUserId := schemas.UserId(parts[1])

	// 2. Leggi il Body (chi è il destinatario?)
	var req struct {
		RecipientId schemas.UserId `json:"recipientId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 3. Validazione base
	if req.RecipientId == "" {
		http.Error(w, "RecipientId is required", http.StatusBadRequest)
		return
	}
	// Non puoi parlare da solo (opzionale, ma ha senso)
	if req.RecipientId == myUserId {
		http.Error(w, "You cannot chat with yourself", http.StatusBadRequest)
		return
	}

	// 4. Chiama il DB
	conversation, err := rt.db.CreateConversation(myUserId, req.RecipientId)
	if err != nil {
		// Se l'errore è grave logghiamo, altrimenti gestiamo i casi (es. utente B non esiste)
		rt.baseLogger.WithError(err).Error("Error creating conversation")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Rispondi con la conversazione (200 OK se esiste, 201 se creata...
	// la specifica dice 200 per entrambe nel caso idempotente, va bene 200)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversation)
}
