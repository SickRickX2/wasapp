package database

import (
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) CreateMessage(convId schemas.ConversationId, msg schemas.Message) error {
	// 1. Controllo: Il mittente è un partecipante?
	const checkQuery = `SELECT 1 FROM conversation_participants WHERE convId = ? AND userId = ?`
	var found int
	err := db.c.QueryRow(checkQuery, convId, msg.Sender).Scan(&found)
	if err != nil {
		return errors.New("user not authorized to send message in this conversation")
	}

	// 2. Inserimento
	const sqlQuery = `
		INSERT INTO messages (messageId, convId, senderId, text, sentAt, kind, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = db.c.Exec(sqlQuery,
		msg.ID,
		convId,
		msg.Sender,
		msg.Text,
		msg.Time,
		"normal", // kind
		"sent",   // status
	)
	return err
}
