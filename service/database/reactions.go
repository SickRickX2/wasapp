package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) ReactToMessage(messageId schemas.MessageId, userId schemas.UserId, emoji string) error {
	// INSERT OR REPLACE: Se la riga (messageId, userId) esiste già, sostituisce l'emoji.
	// Se non esiste, la crea. Perfetto per la logica "una reazione per utente".
	const query = `
		INSERT OR REPLACE INTO message_reactions (messageId, userId, emoji)
		VALUES (?, ?, ?)
	`
	_, err := db.c.Exec(query, messageId, userId, emoji)
	return err
}

func (db *appdbimpl) UnreactToMessage(messageId schemas.MessageId, userId schemas.UserId) error {
	const query = `DELETE FROM message_reactions WHERE messageId = ? AND userId = ?`
	_, err := db.c.Exec(query, messageId, userId)
	return err
}
