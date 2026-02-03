package api

import (
	"encoding/json"
	"net/http"
)

func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request) {
	//  legge la query
	query := r.URL.Query().Get("q")

	// validazione
	if len(query) < 1 {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

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
