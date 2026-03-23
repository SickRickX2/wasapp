package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Autenticazione (Estraiamo il VERO utente da qui!)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// QUESTO è il "me" di cui parla l'URL
	myUserId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))

	// 2. Prende convId dal path (userId non c'è più nell'URL!)
	convId := schemas.ConversationId(ps.ByName("convId"))

	// 3. Rimuove il membro dal DB usando myUserId
	// Prima salviamo un messaggio di sistema persistente finché l'utente è ancora membro
	leavingUserLabel := string(myUserId)
	if leavingUser, userErr := rt.db.GetUserById(myUserId); userErr == nil && leavingUser.Name != "" {
		leavingUserLabel = leavingUser.Name
	}

	systemMsg := schemas.Message{
		ID:     generateMessageId(),
		Sender: myUserId,
		Status: schemas.MsgStatusSent,
		Kind:   "system_leave_group",
		Time:   time.Now(),
		Text:   leavingUserLabel,
	}
	if createErr := rt.db.CreateMessage(convId, systemMsg); createErr != nil {
		rt.baseLogger.WithError(createErr).Warn("failed to create system leave-group message")
	}

	err := rt.db.RemoveGroupMember(convId, myUserId)
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

	// 4. RECUPERA IL GRUPPO AGGIORNATO (Requisito YAML!)
	// Nota: Assicurati di avere una funzione GetConversation o GetGroup nel DB
	updatedGroup, err := rt.db.GetConversation(convId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error fetching updated group")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	group, ok := updatedGroup.(schemas.Group)
	if !ok {
		rt.baseLogger.Error("updated conversation is not a group in leaveGroup")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Risposta 200 OK con JSON (E accontentiamo errcheck)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(group); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in leaveGroup")
	}
}
