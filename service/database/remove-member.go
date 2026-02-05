package database

import (
	"database/sql"
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) RemoveGroupMember(convId schemas.ConversationId, userId schemas.UserId) error {
	// controlla se è un gruppo
	const checkQuery = `SELECT kind FROM conversations WHERE convId = ?`
	var kind string
	err := db.c.QueryRow(checkQuery, convId).Scan(&kind)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("conversation not found")
		}
		return err
	}
	if kind != "group" {
		return errors.New("cannot remove members from a private conversation")
	}

	// toglie l'utente
	const deleteQuery = `DELETE FROM conversation_participants WHERE convId = ? AND userId = ?`
	res, err := db.c.Exec(deleteQuery, convId, userId)
	if err != nil {
		return err
	}

	// controlla se è stato effettivamente rimosso
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("user not in group")
	}

	return nil
}
