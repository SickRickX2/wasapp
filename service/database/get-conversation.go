package database

import (
	"database/sql"
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

// GetConversation restituisce "any" perché può essere una PrivateConversation o un Group
func (db *appdbimpl) GetConversation(convId schemas.ConversationId) (any, error) {
	// 1. Variabili per raccogliere i dati dal DB
	var kind string
	var groupName sql.NullString
	var groupPhoto sql.NullString
	var createdAt sql.NullTime // Usiamo NullTime perché le private potrebbero non averlo nel DB

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

	// 2. Estrae la lista dei partecipanti
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

	// 3. IL BIVIO POLIMORFICO: Costruiamo la struct giusta!
	if kind == groupType {
		// Restituiamo la struct Group (che HA il campo GroupPhoto!)
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
		// Restituiamo la struct PrivateConversation
		privateChat := schemas.PrivateConversation{
			ConvId:       convId,
			Type:         "private",
			Participants: participants,
			UnreadCount:  0, // TODO: da implementare in seguito
		}
		return privateChat, nil
	}
}
