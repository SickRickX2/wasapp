package database

import (
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) AddGroupMembers(convId schemas.ConversationId, userIds []schemas.UserId) error {
	// verifica che la conversazione sia un gruppo
	const checkQuery = `SELECT kind FROM conversations WHERE convId = ?`
	var kind string
	err := db.c.QueryRow(checkQuery, convId).Scan(&kind)
	if err != nil {
		return err
	}
	if kind != groupType {
		return errors.New("cannot add members to a private conversation")
	}

	// aggiungi gli utenti
	const insertQuery = `INSERT OR IGNORE INTO conversation_participants (convId, userId) VALUES (?, ?)`

	// prepariamo lo statement per non ricompilare la query N volte
	stmt, err := db.c.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()
	// se fallisce fermiamo
	for _, userId := range userIds {
		_, err = stmt.Exec(convId, userId)
		if err != nil {
			return err
		}
	}

	return nil
}
