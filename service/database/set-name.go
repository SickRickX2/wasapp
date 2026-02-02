package database

import "github.com/SickRickX2/wasapp/service/api/schemas"

func (db *appdbimpl) SetUserName(id schemas.UserId, newName string) error {
	const query = `UPDATE users SET userName = ? WHERE userId = ?`
	_, err := db.c.Exec(query, newName, id)
	return err
}
