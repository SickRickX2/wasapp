package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) MarkAsSeen(convId schemas.ConversationId, messageId schemas.MessageId, readerId schemas.UserId) error {
	// 1. Troviamo la data del messaggio target
	// Serve per la logica "implicitly previous ones"
	var sentAt time.Time
	err := db.c.QueryRow("SELECT sentAt FROM messages WHERE messageId = ? AND convId = ?", messageId, convId).Scan(&sentAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("message not found")
		}
		return err
	}

	// 2. Aggiorniamo lo stato a 'seen' per:
	// - Messaggi in questa chat
	// - Inviati PRIMA o NELLO STESSO MOMENTO del messaggio target
	// - Che NON sono stati inviati dall'utente che sta leggendo (readerId)
	// - Che non sono già 'seen' (per ottimizzare)
	const updateQuery = `
		UPDATE messages 
		SET status = 'seen' 
		WHERE convId = ? 
		  AND sentAt <= ? 
		  AND senderId != ? 
		  AND status != 'seen'
	`
	_, err = db.c.Exec(updateQuery, convId, sentAt, readerId)
	return err
}
