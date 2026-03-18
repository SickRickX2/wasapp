package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// prende l'utente
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized: missing header", http.StatusUnauthorized)
		return
	}
	// parsa il token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != bearerPrefix {
		http.Error(w, "Unauthorized: invalid token format", http.StatusUnauthorized)
		return
	}
	myUserId := schemas.UserId(parts[1])

	// cerca il destinatario nel body
	var req struct {
		RecipientId schemas.UserId `json:"recipientId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// validazioni
	if req.RecipientId == "" {
		http.Error(w, "RecipientId is required", http.StatusBadRequest)
		return
	}
	// non puoi parlare da solo in una conversazione privata
	if req.RecipientId == myUserId {
		http.Error(w, "You cannot chat with yourself", http.StatusBadRequest)
		return
	}

	// crea se none siste
	conversation, err := rt.db.CreateConversation(myUserId, req.RecipientId)
	if err != nil {
		// problemi con il database
		rt.baseLogger.WithError(err).Error("Error creating conversation")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// risponde con la conversazione
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversation)
}
