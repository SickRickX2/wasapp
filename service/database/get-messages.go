package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) GetConversationMessages(convId schemas.ConversationId, limit int, beforeId string) ([]schemas.Message, error) {
	var messages []schemas.Message

	// prendiamo la conversazione
	query := `SELECT messageId, senderId, text, sentAt, kind, status FROM messages WHERE convId = ?`
	args := []interface{}{convId}

	// prendiamo i messaggi prima di beforeid
	if beforeId != "" {
		query += ` AND sentAt < (SELECT sentAt FROM messages WHERE messageId = ?)`
		args = append(args, beforeId)
	}

	// ordiniamo
	query += ` ORDER BY sentAt DESC LIMIT ?`
	args = append(args, limit)

	rows, err := db.c.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m schemas.Message
		// ingnora i media
		if err := rows.Scan(&m.ID, &m.Sender, &m.Text, &m.Time, &m.Kind, &m.Status); err != nil {
			return nil, err
		}
		// mette le reazioni vuote
		m.Reactions = []schemas.Reaction{}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// non deve restituire nil
	if messages == nil {
		messages = make([]schemas.Message, 0)
	}

	return messages, nil
}
