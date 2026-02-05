package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/gorilla/mux"
)

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// DE-COMMENTATO: Ora ci serve sapere chi è l'utente!
	userId := schemas.UserId(strings.TrimPrefix(authHeader, "Bearer "))

	// 2. Parametri Path
	vars := mux.Vars(r)
	convId := schemas.ConversationId(vars["convId"])

	// 3. Controllo Permessi: L'utente fa parte del gruppo?
	isInGroup, err := rt.db.IsUserInConversation(convId, userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error checking group membership")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isInGroup {
		// Se non è nel gruppo, 403 Forbidden!
		http.Error(w, "Forbidden: You are not a participant of this group", http.StatusForbidden)
		return
	}

	// 4. Parsing Body
	var req struct {
		MediaId string `json:"mediaId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.MediaId == "" {
		http.Error(w, "mediaId is required", http.StatusBadRequest)
		return
	}

	// Costruiamo l'URL
	photoUrl := "/images/" + req.MediaId + ".jpg"

	// 5. Aggiorna DB
	err = rt.db.SetGroupPhoto(convId, photoUrl)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error setting group photo")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
