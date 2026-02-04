package database

import (
	"database/sql"
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) RemoveGroupMember(convId schemas.ConversationId, userId schemas.UserId) error {
	// 1. Verifica che sia un GRUPPO
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

	// 2. Rimuovi l'utente dalla tabella partecipanti
	const deleteQuery = `DELETE FROM conversation_participants WHERE convId = ? AND userId = ?`
	res, err := db.c.Exec(deleteQuery, convId, userId)
	if err != nil {
		return err
	}

	// Opzionale: controlla se l'abbiamo rimosso davvero
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("user not in group")
	}

	return nil
}
