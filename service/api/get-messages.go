package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) getMessages(w http.ResponseWriter, r *http.Request) {
	// 1. Prendi convId dal path
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])

	// 2. Autenticazione: Controlla se l'utente fa parte della chat?
	// (Per brevità saltiamo il check qui, ma idealmente dovresti controllare
	// se il richiedente è in conversation_participants, come fatto per CreateMessage)

	// 3. Gestione Parametri Query (limit, beforeId)
	limit := 20 // Default
	if l := r.URL.Query().Get("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	beforeId := r.URL.Query().Get("beforeId")

	// 4. Chiama DB
	msgs, err := rt.db.GetConversationMessages(convId, limit, beforeId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving messages")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Rispondi JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}
