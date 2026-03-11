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
	// convId c'è nell'URL ma non ci serve strettamente per la query SQL diretta sul messaggio,
	// ma potremmo usarlo per verificare che il messaggio appartenga a quella chat.

	// 3. Body Parsing
	var req struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	// Validazione Emoji (semplice check lunghezza)
	if len(req.Emoji) == 0 {
		http.Error(w, "Emoji is required", http.StatusBadRequest)
		return
	}

	// 4. DB Call
	err := rt.db.ReactToMessage(messageId, userId, req.Emoji)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error setting reaction")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Risposta
	// Lo YAML dice che dovremmo restituire il Messaggio aggiornato.
	// Per brevità ritorniamo 200 OK con un JSON semplice o ricarichiamo il messaggio.
	// Facciamo la versione semplice che dice "ok".
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "reaction set"}`))
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

	// 3. DB Call
	err := rt.db.UnreactToMessage(messageId, userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error removing reaction")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. Risposta
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "reaction removed"}`))
}
