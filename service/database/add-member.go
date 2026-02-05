package database

import (
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) AddGroupMembers(convId schemas.ConversationId, userIds []schemas.UserId) error {
	// 1. Verifica che la conversazione sia un GRUPPO (lo facciamo una volta sola)
	const checkQuery = `SELECT kind FROM conversations WHERE convId = ?`
	var kind string
	err := db.c.QueryRow(checkQuery, convId).Scan(&kind)
	if err != nil {
		return err // Conversazione non trovata o errore DB
	}
	if kind != "group" {
		return errors.New("cannot add members to a private conversation")
	}

	// 2. Aggiungi gli utenti
	// Nota: Potremmo usare una transazione qui, ma per semplicità facciamo un loop.
	// Se un utente è già dentro, INSERT OR IGNORE lo salta.
	const insertQuery = `INSERT OR IGNORE INTO conversation_participants (convId, userId) VALUES (?, ?)`

	// Prepariamo lo statement per non ricompilare la query N volte
	stmt, err := db.c.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, userId := range userIds {
		_, err = stmt.Exec(convId, userId)
		if err != nil {
			return err // Se fallisce l'inserimento (es. DB down), ci fermiamo
		}
	}

	return nil
}
