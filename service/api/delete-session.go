package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) logout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.WriteHeader(http.StatusNoContent)
}
