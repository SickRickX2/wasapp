package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type loginRequest struct {
	Name string `json:"name"`
}

type loginResponse struct {
	Identifier string `json:"identifier"`
}

func (rt *_router) postSession(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if len(body.Name) < 3 || len(body.Name) > 16 {
		http.Error(w, "invalid name length", http.StatusBadRequest)
		return
	}
	resp := loginResponse{Identifier: "usr_demo123"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
