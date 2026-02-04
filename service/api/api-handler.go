package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	//rt.router.GET("/", rt.getHelloWorld)
	rt.router.HandleFunc("/session", rt.doLogin).Methods("POST")
	//rt.router.GET("/context", rt.wrap(rt.getContextReply))
	// Special routes
	//rt.router.GET("/liveness", rt.liveness)
	//rt.router.HandleFunc("/users/{userId}/username", rt.setUsername).Methods("PUT")
	rt.router.HandleFunc("/liveness", rt.liveness).Methods("GET")
	rt.router.HandleFunc("/users/{userId}/username", rt.setUsername).Methods("PUT")
	rt.router.HandleFunc("/users", rt.searchUsers).Methods("GET")
	rt.router.HandleFunc("/conversations", rt.createConversation).Methods("POST")
	rt.router.HandleFunc("/conversations/{convId}/messages", rt.sendMessage).Methods("POST")
	rt.router.HandleFunc("/conversations/{convId}/messages", rt.getMessages).Methods("GET")
	rt.router.HandleFunc("/conversations/{convId}/messages/{messageId}", rt.deleteMessage).Methods("DELETE")
	rt.router.HandleFunc("/groups", rt.createGroup).Methods("POST")
	rt.router.HandleFunc("/conversations/{convId}/participants/{userId}", rt.addToGroup).Methods("PUT")

	return rt.router
}
