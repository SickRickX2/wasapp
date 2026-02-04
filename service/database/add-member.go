package database

import (
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) AddGroupMember(convId schemas.ConversationId, userId schemas.UserId) error {
	// 1. Verifica che la conversazione sia un GRUPPO
	// (Non vogliamo aggiungere gente a chat private 1-vs-1)
	const checkQuery = `SELECT kind FROM conversations WHERE convId = ?`
	var kind string
	err := db.c.QueryRow(checkQuery, convId).Scan(&kind)
	if err != nil {
		return err // Conversazione non trovata
	}
	if kind != "group" {
		return errors.New("cannot add members to a private conversation")
	}

	// 2. Aggiungi l'utente
	// Usiamo INSERT OR IGNORE (o gestiamo l'errore) per evitare doppioni se l'utente è già dentro
	const insertQuery = `INSERT OR IGNORE INTO conversation_participants (convId, userId) VALUES (?, ?)`
	_, err = db.c.Exec(insertQuery, convId, userId)
	return err
}
