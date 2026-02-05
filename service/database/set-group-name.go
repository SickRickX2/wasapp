package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) SetGroupName(convId schemas.ConversationId, newName string) error {
	// Aggiorna il nome solo se è un gruppo (kind='group')
	const query = `UPDATE conversations SET groupName = ? WHERE convId = ? AND kind = 'group'`
	_, err := db.c.Exec(query, newName, convId)
	return err
}
