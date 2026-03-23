package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != bearerPrefix {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	creatorId := schemas.UserId(parts[1])

	// parsa il body
	var req struct {
		GroupName    string           `json:"groupName"`
		Participants []schemas.UserId `json:"participants"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// validazione
	if len(req.GroupName) < 1 || len(req.GroupName) > 30 {
		http.Error(w, "Group name must be between 1 and 30 chars", http.StatusBadRequest)
		return
	}
	// forse dovrei controlalre  che la lista non sia vuota ma non è obbligatorio

	// crea il gruppo
	group, err := rt.db.CreateGroup(creatorId, req.GroupName, req.Participants)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error creating group")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// messaggio di sistema persistente: gruppo creato
	systemMsg := schemas.Message{
		ID:     generateMessageId(),
		Sender: creatorId,
		Status: schemas.MsgStatusSent,
		Kind:   "system_group_created",
		Time:   time.Now(),
		Text:   req.GroupName,
	}
	if err := rt.db.CreateMessage(group.ConvId, systemMsg); err != nil {
		rt.baseLogger.WithError(err).Warn("failed to create system message for group creation")
	}

	// risposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(group); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in createGroup")
	}
}
