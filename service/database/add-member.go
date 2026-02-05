package database

import (
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) AddGroupMember(convId schemas.ConversationId, userId schemas.UserId) error {
	// verifica che il gruppo esiste
	const checkQuery = `SELECT kind FROM conversations WHERE convId = ?`
	var kind string
	err := db.c.QueryRow(checkQuery, convId).Scan(&kind)
	if err != nil {
		return err // Conversazione non trovata
	}
	if kind != "group" {
		return errors.New("cannot add members to a private conversation")
	}

	// aggiunge il membro se nonci sta già (INSERT OR IGNORE)
	const insertQuery = `INSERT OR IGNORE INTO conversation_participants (convId, userId) VALUES (?, ?)`
	_, err = db.c.Exec(insertQuery, convId, userId)
	return err
}
