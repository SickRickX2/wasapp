package database

import (
	"database/sql"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) GetConversations(userId schemas.UserId) ([]schemas.Conversation, error) {

	var convs []schemas.Conversation

	// prende tutte le conversazioni in cui partecipa
	const query = `
		SELECT c.convId, c.kind, c.groupName, c.groupPhoto, c.createdAt
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
		var groupName sql.NullString
		var groupPhoto sql.NullString

		err := rows.Scan(&c.ConvId, &c.Type, &groupName, &groupPhoto, &c.CreatedAt)
		if err != nil {
			return nil, err
		}

		if c.Type == groupType {
			if groupName.Valid {
				c.GroupName = groupName.String
			}
			if groupPhoto.Valid {
				c.GroupPhoto = groupPhoto.String
			}
		}

		const participantsQuery = `
			SELECT cp.userId, COALESCE(u.userName, '')
			FROM conversation_participants cp
			LEFT JOIN users u ON u.userId = cp.userId
			WHERE cp.convId = ?
		`
		participantRows, err := db.c.Query(participantsQuery, c.ConvId)
		if err != nil {
			return nil, err
		}

		participants := make([]schemas.UserId, 0)
		participantNames := make([]string, 0)
		for participantRows.Next() {
			var participantId schemas.UserId
			var participantName string
			if err := participantRows.Scan(&participantId, &participantName); err != nil {
				_ = participantRows.Close()
				return nil, err
			}
			participants = append(participants, participantId)
			participantNames = append(participantNames, participantName)
		}
		if err := participantRows.Err(); err != nil {
			_ = participantRows.Close()
			return nil, err
		}
		if err := participantRows.Close(); err != nil {
			return nil, err
		}

		c.Participants = participants
		c.ParticipantNames = participantNames

		lastMessage, unreadCount, err := db.loadConversationSummary(c.ConvId, userId)
		if err != nil {
			return nil, err
		}
		c.LastMessage = lastMessage
		c.UnreadCount = unreadCount
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
