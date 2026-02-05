package database

import (
	"database/sql"
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) CreateMessage(convId schemas.ConversationId, msg schemas.Message) error {

	// 1. Controllo se l'utente fa parte della conversazione
	const checkQuery = `SELECT 1 FROM conversation_participants WHERE convId = ? AND userId = ?`
	var found int
	err := db.c.QueryRow(checkQuery, convId, msg.Sender).Scan(&found)
	if err != nil {
		return errors.New("user not authorized to send message in this conversation")
	}

	// 2. Gestione Media
	var mediaId sql.NullString
	if msg.MediaId != "" {
		mediaId.String = msg.MediaId
		mediaId.Valid = true
	} else {
		mediaId.Valid = false
	}

	// 3. Gestione ReplyToId
	var replyToId sql.NullString
	var kind string = "normal"

	// 4. Inserimento nel DB
	const sqlQuery = `
		INSERT INTO messages (messageId, convId, senderId, text, mediaId, replyToId, sentAt, kind, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = db.c.Exec(sqlQuery,
		msg.ID,
		convId,
		msg.Sender,
		msg.Text,
		mediaId,
		replyToId,
		msg.Time,
		kind,
		"sent",
	)

	return err
}
