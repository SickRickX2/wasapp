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

	// Variabile di supporto per leggere la data dal DB (anche se non la usiamo nel JSON finale)
	var createdAt time.Time

	// 1. Controlla se esiste già una chat PRIVATA tra questi due utenti
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
		// Trovata! Restituisci quella esistente
		conv.Type = "private"
		conv.Participants = []schemas.UserId{userA, userB}
		return conv, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return conv, err
	}

	// 2. Se non esiste, creala
	newConvID := generateConvId()
	now := time.Now()

	tx, err := db.c.Begin()
	if err != nil {
		return conv, err
	}
	defer tx.Rollback()

	// A. Insert Conversation
	_, err = tx.Exec(`INSERT INTO conversations (convId, kind, createdAt) VALUES (?, 'private', ?)`, newConvID, now)
	if err != nil {
		return conv, err
	}

	// B. Insert Participants
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

	// 3. Costruisci l'oggetto (SENZA CreatedAt)
	conv = schemas.PrivateConversation{
		ConvId:       newConvID,
		Type:         "private",
		Participants: []schemas.UserId{userA, userB},
		// CreatedAt: now, <--- RIMOSSO perché non esiste nello schema API
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
