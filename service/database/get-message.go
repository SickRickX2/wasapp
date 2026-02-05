package database

import (
	"database/sql"
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) GetMessage(messageId schemas.MessageId) (schemas.Message, error) {
	var msg schemas.Message
	var mediaId sql.NullString // Per gestire il possibile NULL del media

	// Recuperiamo i dati essenziali da copiare
	const query = `
		SELECT messageId, senderId, text, mediaId, kind, sentAt
		FROM messages
		WHERE messageId = ?
	`
	err := db.c.QueryRow(query, messageId).Scan(
		&msg.ID,
		&msg.Sender,
		&msg.Text,
		&mediaId,
		&msg.Kind,
		&msg.Time,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return msg, errors.New("message not found")
		}
		return msg, err
	}

	// Convertiamo NullString in stringa normale
	if mediaId.Valid {
		msg.MediaId = mediaId.String
	}

	return msg, nil
}
