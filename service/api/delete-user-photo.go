package api

import (
	"net/http"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

// DELETE /users/{userId}/pfp
func (rt *_router) deleteMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	requestingUserId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))
	targetUserId := schemas.UserId(ps.ByName("userId"))

	if targetUserId == "" {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	if requestingUserId != targetUserId {
		http.Error(w, "You can only delete your own photo", http.StatusForbidden)
		return
	}

	if err := rt.db.SetUserPhoto(targetUserId, ""); err != nil {
		rt.baseLogger.WithError(err).Error("Error deleting user photo")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
