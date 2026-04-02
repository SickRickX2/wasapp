package database

import (
	"database/sql"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) loadConversationSummary(convId schemas.ConversationId, userId schemas.UserId) (*schemas.Message, int, error) {
	var lastMessage *schemas.Message
	{
		const query = `
			SELECT messageId, senderId, text, mediaId, replyToId, sentAt, kind, status
			FROM messages
			WHERE convId = ?
			ORDER BY sentAt DESC
			LIMIT 1
		`

		row := db.c.QueryRow(query, convId)
		var m schemas.Message
		var mediaIdSql sql.NullString
		var replyToIdSql sql.NullString
		if err := row.Scan(&m.ID, &m.Sender, &m.Text, &mediaIdSql, &replyToIdSql, &m.Time, &m.Kind, &m.Status); err != nil {
			if err != sql.ErrNoRows {
				return nil, 0, err
			}
		} else {
			if mediaIdSql.Valid {
				m.MediaId = mediaIdSql.String
			}
			if replyToIdSql.Valid {
				val := schemas.MessageId(replyToIdSql.String)
				m.ReplyToId = &val
			}
			lastMessage = &m
		}
	}

	var unreadCount int
	{
		if userId == "" {
			return lastMessage, 0, nil
		}

		const query = `
			SELECT COUNT(*)
			FROM messages
			WHERE convId = ?
			  AND senderId != ?
			  AND status != 'seen'
			  AND status != 'deleted'
		`
		if err := db.c.QueryRow(query, convId, userId).Scan(&unreadCount); err != nil {
			return nil, 0, err
		}
	}

	return lastMessage, unreadCount, nil
}
