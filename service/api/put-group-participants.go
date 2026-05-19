package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	appErrs "github.com/SickRickX2/wasapp/service/api/errs"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	requestingUserId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))

	vars := ps
	convId := schemas.ConversationId(vars.ByName("convId"))
	if convId == "" {
		http.Error(w, "Invalid convId", http.StatusBadRequest)
		return
	}

	// solo i partecipanti possono aggiungere
	isInConversation, err := rt.db.IsUserInConversation(convId, requestingUserId)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !isInConversation {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	//parsa il body
	var req struct {
		UserIds []schemas.UserId `json:"userIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// almeno 1 utente
	if len(req.UserIds) == 0 {
		http.Error(w, "User list cannot be empty", http.StatusBadRequest)
		return
	}
	if len(req.UserIds) > 20 {
		http.Error(w, "Too many users", http.StatusBadRequest)
		return
	}

	// aggiunge al db
	err = rt.db.AddGroupMembers(convId, req.UserIds)
	if err != nil {
		switch {
		case errors.Is(err, appErrs.ErrCannotAddMembersToPrivate):
			http.Error(w, "Cannot add members to a private conversation", http.StatusBadRequest)
		case errors.Is(err, sql.ErrNoRows):
			http.Error(w, "Conversation not found", http.StatusNotFound)
		default:
			rt.baseLogger.WithError(err).Error("Error adding members to group")
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		return
	}

	// messaggi di sistema persistenti per ogni utente aggiunto
	for _, addedUserId := range req.UserIds {
		addedUserLabel := string(addedUserId)
		if addedUser, userErr := rt.db.GetUserById(addedUserId); userErr == nil && addedUser.Name != "" {
			addedUserLabel = addedUser.Name
		}

		systemMsg := schemas.Message{
			ID:     generateMessageId(),
			Sender: requestingUserId,
			Status: schemas.MsgStatusSent,
			Kind:   "system_add_member",
			Time:   time.Now(),
			Text:   addedUserLabel,
		}

		if createErr := rt.db.CreateMessage(convId, systemMsg); createErr != nil {
			rt.baseLogger.WithError(createErr).Warn("failed to create system add-member message")
		}
	}

	// risposta con il gruppo aggiornato
	updatedConv, err := rt.db.GetConversation(convId)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}
	updatedGroup, ok := updatedConv.(schemas.Group)
	if !ok {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedGroup); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in addToGroup")
	}
}
