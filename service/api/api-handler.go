package api

import (
	"net/http"
)

const bearerPrefix = "Bearer"

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// --- Sessione ---
	rt.router.POST("/session", rt.doLogin)
	rt.router.DELETE("/session", rt.logout)

	// liveness
	rt.router.GET("/liveness", rt.liveness)

	// --- users ---
	rt.router.GET("/users", rt.searchUsers)
	rt.router.PUT("/users/:userId/username", rt.setUsername)
	rt.router.PUT("/users/:userId/pfp", rt.setUserPhoto)

	// --- conversations ---
	rt.router.GET("/conversations", rt.getConversations)
	rt.router.POST("/conversations", rt.createConversation)
	rt.router.POST("/groups", rt.createGroup)
	rt.router.POST("/conversations/:convId/participants", rt.addToGroup)
	rt.router.DELETE("/conversations/:convId/participants/:userId", rt.removeFromGroup)
	rt.router.PUT("/conversations/:convId/group_photo", rt.setGroupPhoto)
	rt.router.PUT("/conversations/:convId/group_name", rt.setGroupName)

	// --- messages ---
	rt.router.GET("/conversations/:convId/messages", rt.getMessages)
	rt.router.POST("/conversations/:convId/messages", rt.sendMessage)
	rt.router.DELETE("/conversations/:convId/messages/:messageId", rt.deleteMessage)
	rt.router.POST("/conversations/:convId/messages/:messageId/forwarded", rt.forwardMessage)
	rt.router.PUT("/conversations/:convId/messages/:messageId/seen", rt.markAsSeen)

	// --- reactions ---
	rt.router.PUT("/conversations/:convId/messages/:messageId/reaction", rt.commentMessage)
	rt.router.DELETE("/conversations/:convId/messages/:messageId/reaction", rt.uncommentMessage)

	// --- media ---
	rt.router.POST("/media", rt.uploadMedia)

	// questo mi serve per accedere alle immagini caricate
	rt.router.ServeFiles("/images/*filepath", http.Dir("./images"))

	return rt.router
}
