package api

import (
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request) {
	// autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// prende i parametri
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])
	targetUserId := schemas.UserId(vars["userId"]) // L'utente da aggiungere

	// aggiunge il membro al gruppo
	err := rt.db.AddGroupMember(convId, targetUserId)
	if err != nil {
		if err.Error() == "cannot add members to a private conversation" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			rt.baseLogger.WithError(err).Error("Error adding member to group")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// risposta
	w.WriteHeader(http.StatusNoContent)
}
