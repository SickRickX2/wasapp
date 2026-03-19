package database

import (
	"database/sql"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) IsUserInConversation(convId schemas.ConversationId, userId schemas.UserId) (bool, error) {
	var dummy int
	// cerchiamo se esiste una riga nella tabella partecipanti
	err := db.c.QueryRow(`SELECT 1 FROM conversation_participants WHERE convId = ? AND userId = ?`, convId, userId).Scan(&dummy)

	if err == sql.ErrNoRows {
		// non è nel gruppo
		return false, nil
	}
	if err != nil {
		// errore del db
		return false, err
	}

	// tutto ok
	return true, nil
}
