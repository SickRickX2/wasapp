package api

import (
	"context"
)

type authKey string

const contextKeyUserID authKey = "userID"

// GetUserID estrae l'ID utente dal contesto.
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextKeyUserID).(string)
	return userID, ok
}
