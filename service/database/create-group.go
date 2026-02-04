package database

import (
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) CreateGroup(creator schemas.UserId, name string, participants []schemas.UserId) (schemas.Group, error) {
	var group schemas.Group

	// Generiamo ID e data
	newConvID := generateConvId() // Assicurati che questa funzione helper sia visibile (es. in create-conversation.go)
	now := time.Now()

	// Inizia Transazione
	tx, err := db.c.Begin()
	if err != nil {
		return group, err
	}
	defer tx.Rollback()

	// 1. Inserisci il Gruppo in 'conversations'
	// Nel DB la colonna si chiama 'kind', ma nello schema JSON è 'type'
	const insertConv = `
		INSERT INTO conversations (convId, kind, groupName, createdAt) 
		VALUES (?, 'group', ?, ?)
	`
	_, err = tx.Exec(insertConv, newConvID, name, now)
	if err != nil {
		return group, err
	}

	// 2. Aggiungi i Partecipanti
	// Usiamo una map per evitare duplicati e includere sempre il creatore
	uniqueUsers := make(map[schemas.UserId]bool)
	uniqueUsers[creator] = true
	for _, u := range participants {
		uniqueUsers[u] = true
	}

	stmt, err := tx.Prepare(`INSERT INTO conversation_participants (convId, userId) VALUES (?, ?)`)
	if err != nil {
		return group, err
	}
	defer stmt.Close()

	// Inseriamo tutti nel DB e prepariamo la lista per la risposta JSON
	finalParticipants := []schemas.UserId{}
	for u := range uniqueUsers {
		_, err = stmt.Exec(newConvID, u)
		if err != nil {
			return group, err
		}
		finalParticipants = append(finalParticipants, u)
	}

	// 3. Commit
	if err = tx.Commit(); err != nil {
		return group, err
	}

	// 4. Costruisci risposta usando il TUO schemas.go
	group.ConvId = newConvID
	group.Type = "group" // Costante ConvTypeGroup
	group.GroupName = name
	group.Participants = finalParticipants
	group.CreatedAt = now
	group.UnreadCount = 0

	return group, nil
}
