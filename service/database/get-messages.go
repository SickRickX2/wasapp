package database

import (
	"database/sql"
	"sort"

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

		// prende le reazioni per questo messaggio e le raggruppa per emoji
		reacRows, err := db.c.Query(`
			SELECT mr.userId, mr.emoji, IFNULL(u.userName, mr.userId)
			FROM message_reactions mr
			LEFT JOIN users u ON u.userId = mr.userId
			WHERE mr.messageId = ?
			ORDER BY mr.emoji ASC, IFNULL(u.userName, mr.userId) ASC, mr.userId ASC
		`, m.ID)
		if err != nil {
			return nil, err
		}

		type reactionAccumulator struct {
			emoji   string
			users   []string
			userIds []schemas.UserId
		}
		reactionGroups := map[string]*reactionAccumulator{}

		for reacRows.Next() {
			var userId schemas.UserId
			var emoji string
			var userName string
			if err := reacRows.Scan(&userId, &emoji, &userName); err != nil {
				reacRows.Close()
				return nil, err
			}
			group, ok := reactionGroups[emoji]
			if !ok {
				group = &reactionAccumulator{emoji: emoji}
				reactionGroups[emoji] = group
			}
			group.users = append(group.users, userName)
			group.userIds = append(group.userIds, userId)
		}

		if err := reacRows.Err(); err != nil {
			reacRows.Close()
			return nil, err
		}
		reacRows.Close()
		// -----------------------

		if len(reactionGroups) > 0 {
			emojis := make([]string, 0, len(reactionGroups))
			for emoji := range reactionGroups {
				emojis = append(emojis, emoji)
			}
			sort.Strings(emojis)

			m.Reactions = make([]schemas.Reaction, 0, len(emojis))
			for _, emoji := range emojis {
				group := reactionGroups[emoji]
				m.Reactions = append(m.Reactions, schemas.Reaction{
					Emoji:   group.emoji,
					Users:   group.users,
					UserIds: group.userIds,
				})
			}
		} else {
			m.Reactions = []schemas.Reaction{}
		}

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
