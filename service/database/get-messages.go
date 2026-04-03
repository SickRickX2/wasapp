package database

import (
	"database/sql"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) GetConversationMessages(convId schemas.ConversationId, limit int, offset int) ([]schemas.Message, error) {
	var messages []schemas.Message

	// paginazione a blocchi: ultimi messaggi prima (DESC), con LIMIT/OFFSET
	query := `
		SELECT messageId, senderId, text, mediaId, replyToId, sentAt, kind, status
		FROM messages
		WHERE convId = ?
		ORDER BY sentAt DESC
		LIMIT ? OFFSET ?
	`
	args := []interface{}{convId, limit, offset}

	rows, err := db.c.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m schemas.Message
		var mediaIdSql sql.NullString
		var replyToIdSql sql.NullString

		if err := rows.Scan(&m.ID, &m.Sender, &m.Text, &mediaIdSql, &replyToIdSql, &m.Time, &m.Kind, &m.Status); err != nil {
			return nil, err
		}

		// Se c'è un mediaId nel DB, lo copiamo nella struct
		if mediaIdSql.Valid {
			m.MediaId = mediaIdSql.String
		}

		// se cè un reply lo copiamo nella struct
		if replyToIdSql.Valid {
			val := schemas.MessageId(replyToIdSql.String)
			m.ReplyToId = &val
		}

		// prende le reazioni per questo messaggio
		reacRows, err := db.c.Query(`SELECT userId, emoji FROM message_reactions WHERE messageId = ?`, m.ID)
		if err != nil {
			return nil, err
		}

		m.Reactions = []schemas.Reaction{}

		for reacRows.Next() {
			var r schemas.Reaction
			if err := reacRows.Scan(&r.UserId, &r.Emoji); err != nil {
				reacRows.Close()
				return nil, err
			}
			m.Reactions = append(m.Reactions, r)
		}

		if err := reacRows.Err(); err != nil {
			reacRows.Close()
			return nil, err
		}
		reacRows.Close()
		// -----------------------

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if messages == nil {
		messages = make([]schemas.Message, 0)
	}

	return messages, nil
}
