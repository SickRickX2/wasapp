package database

import (
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) CreateGroup(creator schemas.UserId, name string, participants []schemas.UserId) (schemas.Group, error) {
	var group schemas.Group

	// validazione del nome
	newConvID := generateConvId()
	now := time.Now()

	// crea il gruppo nel db
	tx, err := db.c.Begin()
	if err != nil {
		return group, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// crea il gruppo nella tabella conversations
	const insertConv = `
		INSERT INTO conversations (convId, kind, groupName, createdAt) 
		VALUES (?, 'group', ?, ?)
	`
	_, err = tx.Exec(insertConv, newConvID, name, now)
	if err != nil {
		return group, err
	}

	// aggiunge i partecipanti
	uniqueUsers := make(map[schemas.UserId]bool)
	uniqueUsers[creator] = true
	for _, u := range participants {
		uniqueUsers[u] = true
	}

	stmt, err := tx.Prepare(`INSERT INTO conversation_participants (convId, userId) VALUES (?, ?)`)
	if err != nil {
		return group, err
	}
	defer func() {
		_ = stmt.Close()
	}()

	// inserisce i partecipanti nella lista finale
	finalParticipants := []schemas.UserId{}
	for u := range uniqueUsers {
		_, err = stmt.Exec(newConvID, u)
		if err != nil {
			return group, err
		}
		finalParticipants = append(finalParticipants, u)
	}

	// esegue la transazione
	if err = tx.Commit(); err != nil {
		return group, err
	}

	// risposta con la struct del gruppo
	group.ConvId = newConvID
	group.Type = groupType
	group.GroupName = name
	group.Participants = finalParticipants
	group.CreatedAt = now
	group.UnreadCount = 0

	return group, nil
}
