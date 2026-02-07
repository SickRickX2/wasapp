package api

import (
	"net/http"
)

func (rt *_router) logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
