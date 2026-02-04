package api

import (
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticazione (Chi sta facendo l'aggiunta?)
	// In una app reale controlleremmo se chi fa la richiesta è amministratore o membro del gruppo.
	// Per ora controlliamo solo che sia loggato.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parametri dal Path
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])
	targetUserId := schemas.UserId(vars["userId"]) // L'utente da aggiungere

	// 3. Chiama il DB
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

	// 4. Successo (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}
