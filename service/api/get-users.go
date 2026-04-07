package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//  legge la query
	query := r.URL.Query().Get("q")

	users, err := rt.db.SearchUsers(query)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error searching users")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// risposta
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}
