package database

import (
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) DeleteMessage(convId schemas.ConversationId, messageId schemas.MessageId, userId schemas.UserId) (schemas.Message, error) {
	var msg schemas.Message

	// 1. Esegui l'aggiornamento (UPDATE)
	// - Imposta status a 'deleted'
	// - Sovrascrive il testo
	// - Rimuove il media (mediaId = NULL)
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

	// Controlliamo se abbiamo davvero modificato qualcosa
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return msg, err
	}
	if rowsAffected == 0 {
		return msg, errors.New("message not found or user not authorized")
	}

	// 2. Recupera il messaggio aggiornato per restituirlo
	// Nota: mediaId sarà NULL quindi media sarà nil
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
