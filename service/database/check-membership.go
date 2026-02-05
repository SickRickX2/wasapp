package database

import (
	"database/sql"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) IsUserInConversation(convId schemas.ConversationId, userId schemas.UserId) (bool, error) {
	var dummy int
	// Cerchiamo se esiste una riga nella tabella partecipanti
	err := db.c.QueryRow(`SELECT 1 FROM conversation_participants WHERE convId = ? AND userId = ?`, convId, userId).Scan(&dummy)

	if err == sql.ErrNoRows {
		// Non trovato -> Non è nel gruppo
		return false, nil
	}
	if err != nil {
		// Errore vero del DB
		return false, err
	}

	// Trovato -> È nel gruppo
	return true, nil
}
