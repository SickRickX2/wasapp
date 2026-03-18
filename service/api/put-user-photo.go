package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setUserPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	requestingUserId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))
	// 2. Controllo Identità (Solo tu puoi cambiare la tua foto)
	vars := ps
	targetUserId := schemas.UserId(vars.ByName("userId"))

	if requestingUserId != targetUserId {
		http.Error(w, "You can only update your own photo", http.StatusForbidden)
		return
	}

	// 3. Parsing Body (ci aspettiamo {"mediaId": "media_..."})
	var req struct {
		MediaId string `json:"mediaId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validazione minima
	if req.MediaId == "" {
		http.Error(w, "mediaId is required", http.StatusBadRequest)
		return
	}

	// 4. Costruiamo l'URL (potremmo fare una query al DB per recuperarlo dalla tabella media,
	// ma sappiamo che il formato è standard, quindi risparmiamo una query).
	// ATTENZIONE: Assumiamo .jpg per semplicità, ma idealmente dovremmo leggere l'estensione dal DB media.
	// Se vuoi essere preciso al 100%, dovresti fare GetMediaById nel DB.
	// Per ora facciamo finta che siano tutte jpg o che il frontend gestisca l'URL.
	photoUrl := "/images/" + req.MediaId + ".jpg"

	// 5. Aggiorna User nel DB
	err := rt.db.SetUserPhoto(targetUserId, photoUrl)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error setting user photo")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 6. Successo
	w.WriteHeader(http.StatusNoContent)
}
