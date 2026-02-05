package api

import (
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) removeFromGroup(w http.ResponseWriter, r *http.Request) {
	// autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// prende i parametri
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])
	targetUserId := schemas.UserId(vars["userId"])

	//  rimuove il membro
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

	// risposta
	w.WriteHeader(http.StatusNoContent)
}
