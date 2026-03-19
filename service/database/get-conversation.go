package database

import (
	"database/sql"
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) GetConversation(convId schemas.ConversationId) (any, error) {
	var kind string
	var groupName sql.NullString
	var groupPhoto sql.NullString
	var createdAt sql.NullTime

	const query = `
		SELECT kind, groupName, groupPhoto, createdAt 
		FROM conversations 
		WHERE convId = ?
	`
	err := db.c.QueryRow(query, convId).Scan(&kind, &groupName, &groupPhoto, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("conversation not found")
		}
		return nil, err
	}

	// recupera i partecipanti
	const partQuery = `
		SELECT userId 
		FROM conversation_participants 
		WHERE convId = ?
	`
	rows, err := db.c.Query(partQuery, convId)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	participants := make([]schemas.UserId, 0)
	for rows.Next() {
		var participantId schemas.UserId
		if err := rows.Scan(&participantId); err != nil {
			return nil, err
		}
		participants = append(participants, participantId)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// costruiamo la risposta in base alla conv
	if kind == groupType {
		// group
		group := schemas.Group{
			ConvId:       convId,
			Type:         "group",
			Participants: participants,
			UnreadCount:  0, // TODO: da implementare in seguito
		}
		if groupName.Valid {
			group.GroupName = groupName.String
		}
		if groupPhoto.Valid {
			group.GroupPhoto = groupPhoto.String
		}
		if createdAt.Valid {
			group.CreatedAt = createdAt.Time
		}
		return group, nil

	} else {
		// private
		privateChat := schemas.PrivateConversation{
			ConvId:       convId,
			Type:         "private",
			Participants: participants,
			UnreadCount:  0, // TODO: da implementare in seguito
		}
		return privateChat, nil
	}
}
