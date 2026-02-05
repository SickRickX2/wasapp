package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) GetConversations(userId schemas.UserId) ([]schemas.Conversation, error) {

	var convs []schemas.Conversation

	// prende tutte le conversazioni in cui partecipa
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
		var groupName *string

		err := rows.Scan(&c.ConvId, &c.Type, &groupName, &c.CreatedAt)
		if err != nil {
			return nil, err
		}

		if c.Type == "group" && groupName != nil {
			c.GroupName = *groupName
		}

		// TODO: prendere l'ultimo messaggio e contare i mesaggi non letti
		convs = append(convs, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// non deve restituire nil
	if convs == nil {
		convs = make([]schemas.Conversation, 0)
	}

	return convs, nil
}
