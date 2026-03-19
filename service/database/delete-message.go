package database

import (
	"database/sql"
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) DeleteMessage(convId schemas.ConversationId, messageId schemas.MessageId, userId schemas.UserId) (schemas.Message, error) {
	var msg schemas.Message
	var senderId schemas.UserId

	const checkQuery = `SELECT senderId FROM messages WHERE convId = ? AND messageId = ?`
	err := db.c.QueryRow(checkQuery, convId, messageId).Scan(&senderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return msg, ErrMessageNotFound
		}
		return msg, err
	}
	if senderId != userId {
		return msg, ErrForbidden
	}

	//  trasforma il messaggio in uno cancellato
	const updateQuery = `
		UPDATE messages 
		SET status = 'deleted', 
		    text = 'This message has been deleted',
		    mediaId = NULL 
		WHERE convId = ? AND messageId = ? AND senderId = ?
	`
	res, err := db.c.Exec(updateQuery, convId, messageId, userId)
	if err != nil {
		return msg, err
	}

	// controlla modifiche
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return msg, err
	}
	if rowsAffected == 0 {
		return msg, ErrMessageNotFound
	}

	// recupera il messaggio aggiornato
	const selectQuery = `
		SELECT messageId, senderId, text, sentAt, kind, status 
		FROM messages 
		WHERE messageId = ?
	`
	err = db.c.QueryRow(selectQuery, messageId).Scan(
		&msg.ID, &msg.Sender, &msg.Text, &msg.Time, &msg.Kind, &msg.Status,
	)
	if err != nil {
		return msg, err
	}

	msg.Reactions = []schemas.Reaction{}
	return msg, nil
}
