package database

import (
	"database/sql" // <--- Serve per gestire i NULL del database (sql.NullString)

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) GetConversationMessages(convId schemas.ConversationId, limit int, beforeId string) ([]schemas.Message, error) {
	var messages []schemas.Message

	// 1. MODIFICA QUI: Aggiunto 'replyToId' alla SELECT
	query := `SELECT messageId, senderId, text, mediaId, replyToId, sentAt, kind, status FROM messages WHERE convId = ?`
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
		var mediaIdSql sql.NullString
		var replyToIdSql sql.NullString // <--- Variabile per il reply

		// 2. MODIFICA QUI: Aggiunto &replyToIdSql allo Scan
		if err := rows.Scan(&m.ID, &m.Sender, &m.Text, &mediaIdSql, &replyToIdSql, &m.Time, &m.Kind, &m.Status); err != nil {
			return nil, err
		}

		// Se c'è un mediaId nel DB, lo copiamo nella struct
		if mediaIdSql.Valid {
			m.MediaId = mediaIdSql.String
		}

		// 3. MODIFICA QUI: Gestione ReplyToId
		if replyToIdSql.Valid {
			val := schemas.MessageId(replyToIdSql.String)
			m.ReplyToId = &val
		}

		// --- LOGICA REAZIONI ---
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
