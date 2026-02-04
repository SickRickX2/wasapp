package database

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) CreateConversation(userA schemas.UserId, userB schemas.UserId) (schemas.PrivateConversation, error) {
	var conv schemas.PrivateConversation

	var createdAt time.Time

	// controlla se esiste già una chat privata tra questi due
	const checkQuery = `
		SELECT c.convId, c.createdAt 
		FROM conversations c
		JOIN conversation_participants cp1 ON c.convId = cp1.convId
		JOIN conversation_participants cp2 ON c.convId = cp2.convId
		WHERE c.kind = 'private' 
		AND cp1.userId = ? 
		AND cp2.userId = ?
	`

	err := db.c.QueryRow(checkQuery, userA, userB).Scan(&conv.ConvId, &createdAt)
	if err == nil {
		// se la trova la restituisce
		conv.Type = "private"
		conv.Participants = []schemas.UserId{userA, userB}
		return conv, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return conv, err
	}

	// se nonla trova tocca crearla
	newConvID := generateConvId()
	now := time.Now()

	tx, err := db.c.Begin()
	if err != nil {
		return conv, err
	}
	defer tx.Rollback()

	// crea la conversazione
	_, err = tx.Exec(`INSERT INTO conversations (convId, kind, createdAt) VALUES (?, 'private', ?)`, newConvID, now)
	if err != nil {
		return conv, err
	}

	// mette i partecipanti
	_, err = tx.Exec(`INSERT INTO conversation_participants (convId, userId) VALUES (?, ?)`, newConvID, userA)
	if err != nil {
		return conv, err
	}
	_, err = tx.Exec(`INSERT INTO conversation_participants (convId, userId) VALUES (?, ?)`, newConvID, userB)
	if err != nil {
		return conv, err
	}

	if err = tx.Commit(); err != nil {
		return conv, err
	}

	conv = schemas.PrivateConversation{
		ConvId:       newConvID,
		Type:         "private",
		Participants: []schemas.UserId{userA, userB},
	}

	return conv, nil
}

func generateConvId() schemas.ConversationId {
	const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	const length = 10
	b := make([]byte, length)
	for i := range b {
		b[i] = charSet[rand.Intn(len(charSet))]
	}
	return schemas.ConversationId("conv_" + string(b))
}
