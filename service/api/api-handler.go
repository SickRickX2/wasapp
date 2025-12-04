package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	//rt.router.GET("/", rt.getHelloWorld)
	rt.router.HandleFunc("/v1/session", rt.postSession).Methods("POST")
	//rt.router.GET("/context", rt.wrap(rt.getContextReply))
	// Special routes
	//rt.router.GET("/liveness", rt.liveness)
	rt.router.HandleFunc("/liveness", rt.liveness).Methods("GET")
	return rt.router

}
