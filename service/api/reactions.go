package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

// PUT /conversations/{convId}/messages/{messageId}/reaction
func (rt *_router) setReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Auth (Ricordati che in futuro potresti usare la costante bearerPrefix!)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, "Bearer "))

	// 2. Path Params
	vars := ps
	messageId := schemas.MessageId(vars.ByName("messageId"))

	// 3. Body Parsing
	var req struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	// Validazione Emoji (La Bibbia dice: minLength 1, maxLength 4)
	if len(req.Emoji) < 1 || len(req.Emoji) > 4 {
		http.Error(w, "Emoji length must be between 1 and 4", http.StatusBadRequest)
		return
	}

	// 4. DB Call (Aggiorna/Inserisci reazione)
	err := rt.db.ReactToMessage(messageId, userId, req.Emoji)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error setting reaction")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Recupera il messaggio aggiornato dal DB (Come richiesto da api.yaml)
	updatedMsg, err := rt.db.GetMessage(messageId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving updated message")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 6. Risposta
	w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK) // Opzionale: 200 è il default di Go se non metti nulla
	if err := json.NewEncoder(w).Encode(updatedMsg); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in setReaction")
	}
}

// DELETE /conversations/{convId}/messages/{messageId}/reaction
func (rt *_router) removeReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Auth
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, "Bearer "))

	// 2. Path Params
	vars := ps
	messageId := schemas.MessageId(vars.ByName("messageId"))

	// 3. DB Call (Rimuove reazione)
	err := rt.db.UnreactToMessage(messageId, userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error removing reaction")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. Recupera il messaggio aggiornato dal DB (Come richiesto da api.yaml)
	updatedMsg, err := rt.db.GetMessage(messageId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving updated message")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Risposta
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedMsg); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in removeReaction")
	}
}
