package api

import (
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) removeFromGroup(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parametri Path
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])
	targetUserId := schemas.UserId(vars["userId"])

	// 3. Chiama DB
	err := rt.db.RemoveGroupMember(convId, targetUserId)
	if err != nil {
		switch err.Error() {
		case "conversation not found":
			http.Error(w, "Group not found", http.StatusNotFound)
		case "user not in group":
			http.Error(w, "User is not in the group", http.StatusNotFound)
		case "cannot remove members from a private conversation":
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			rt.baseLogger.WithError(err).Error("Error removing member")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// 4. Successo
	w.WriteHeader(http.StatusNoContent)
}
