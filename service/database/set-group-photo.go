package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) SetGroupPhoto(convId schemas.ConversationId, photoUrl string) error {
	// Aggiorniamo il campo groupPhoto della conversazione
	// Assicurati che la colonna nel tuo DB si chiami 'groupPhoto' o aggiustala qui
	const query = `UPDATE conversations SET groupPhoto = ? WHERE convId = ? AND kind = 'group'`
	_, err := db.c.Exec(query, photoUrl, convId)
	return err
}
