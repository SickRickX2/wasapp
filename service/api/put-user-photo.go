package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setUserPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	requestingUserId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))
	// controllo identità
	vars := ps
	targetUserId := schemas.UserId(vars.ByName("userId"))

	if requestingUserId != targetUserId {
		http.Error(w, "You can only update your own photo", http.StatusForbidden)
		return
	}

	// parsing Body
	var req struct {
		MediaId string `json:"mediaId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// validazione minima
	if req.MediaId == "" {
		http.Error(w, "mediaId is required", http.StatusBadRequest)
		return
	}

	photoUrl, err := rt.db.GetMediaUrl(req.MediaId)
	if err != nil {
		// Se c'è un errore (es. sql.ErrNoRows), significa che l'immagine non esiste
		rt.baseLogger.WithError(err).Error("Media non trovato nel database")
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	// aggiorna user nel db
	err = rt.db.SetUserPhoto(targetUserId, photoUrl)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error setting user photo")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// successo
	w.WriteHeader(http.StatusNoContent)
}
