package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) GetConversations(userId schemas.UserId) ([]schemas.Conversation, error) {
	// Definiamo una struct temporanea che unisce i campi di Group e PrivateConversation
	// per poterli restituire in un'unica lista eterogenea
	var convs []schemas.Conversation

	// Selezioniamo le conversazioni a cui l'utente partecipa
	const query = `
		SELECT c.convId, c.kind, c.groupName, c.createdAt
		FROM conversations c
		JOIN conversation_participants cp ON c.convId = cp.convId
		WHERE cp.userId = ?
		ORDER BY c.createdAt DESC
	`

	rows, err := db.c.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c schemas.Conversation
		var groupName *string // Può essere NULL nelle chat private

		err := rows.Scan(&c.ConvId, &c.Type, &groupName, &c.CreatedAt)
		if err != nil {
			return nil, err
		}

		if c.Type == "group" && groupName != nil {
			c.GroupName = *groupName
		}

		// Nota: Per fare le cose fatte bene (come da specifica), qui dovremmo anche:
		// 1. Recuperare l'ultimo messaggio (LastMessage)
		// 2. Contare i messaggi non letti (UnreadCount)
		// Ma per ora lasciamoli vuoti/zero per far funzionare la lista base.

		convs = append(convs, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Se vuoto, ritorna array vuoto non nil
	if convs == nil {
		convs = make([]schemas.Conversation, 0)
	}

	return convs, nil
}
