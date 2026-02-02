package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) CreateUser(u schemas.User) error {
	const query = `
		INSERT INTO users (userId, userName, createdAt, pfpUrl) 
		VALUES (?, ?, ?, ?)
	`
	// campi della struct User
	_, err := db.c.Exec(query, u.ID, u.Name, u.CreatedAt, u.PFPURL)
	return err
}
