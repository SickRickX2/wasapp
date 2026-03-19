package api

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) logout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
