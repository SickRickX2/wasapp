package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

// GET /users/{userId}
func (rt *_router) getUserById(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetUserId := schemas.UserId(ps.ByName("userId"))
	if targetUserId == "" {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	user, err := rt.db.GetUserById(targetUserId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in getUserById")
	}
}
