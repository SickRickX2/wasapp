package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) SetUserPhoto(userId schemas.UserId, photoUrl string) error {
	// aggiorna il campo pfpUrl
	const query = `UPDATE users SET pfpUrl = ? WHERE userId = ?`
	_, err := db.c.Exec(query, photoUrl, userId)
	return err
}
