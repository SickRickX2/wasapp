package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getPhotoUrl(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userId := schemas.UserId(ps.ByName("userId"))
	if userId == "" {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	user, err := rt.db.GetUserById(userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving user in getPhotoUrl")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	resp := struct {
		PfpURL string `json:"pfpUrl"`
	}{
		PfpURL: user.PFPURL,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in getPhotoUrl")
	}
}
