package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))
	vars := ps
	convId := schemas.ConversationId(vars.ByName("convId"))

	// controlla se nel gruppo
	isInGroup, err := rt.db.IsUserInConversation(convId, userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error checking group membership")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isInGroup {
		http.Error(w, "Forbidden: You are not a participant of this group", http.StatusForbidden)
		return
	}
	var req struct {
		GroupName string `json:"groupName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// validazione del nome, vedi api
	if len(req.GroupName) < 1 || len(req.GroupName) > 30 {
		http.Error(w, "Group name must be between 1 and 30 chars", http.StatusBadRequest)
		return
	}

	// aggiorna nel db
	err = rt.db.SetGroupName(convId, req.GroupName)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error setting group name")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	updatedGroup := schemas.Group{
		ConvId:    convId,
		Type:      "group",
		GroupName: req.GroupName,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updatedGroup); err != nil {
		rt.baseLogger.WithError(err).Error("Error encoding response")
		return
	}
}
