package database

import (
	"database/sql"
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) GetMessage(messageId schemas.MessageId) (schemas.Message, error) {
	var msg schemas.Message
	var mediaId sql.NullString
	var replyToId sql.NullString // <--- Aggiunto

	// Recuperiamo i dati essenziali
	// MODIFICA QUI: Aggiunto replyToId alla query
	const query = `
		SELECT messageId, senderId, text, mediaId, replyToId, kind, sentAt
		FROM messages
		WHERE messageId = ?
	`
	err := db.c.QueryRow(query, messageId).Scan(
		&msg.ID,
		&msg.Sender,
		&msg.Text,
		&mediaId,
		&replyToId, // <--- Scan anche qui
		&msg.Kind,
		&msg.Time,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return msg, errors.New("message not found")
		}
		return msg, err
	}

	if mediaId.Valid {
		msg.MediaId = mediaId.String
	}

	// Se serve sapere a chi rispondeva
	if replyToId.Valid {
		val := schemas.MessageId(replyToId.String)
		msg.ReplyToId = &val
	}

	return msg, nil
}
