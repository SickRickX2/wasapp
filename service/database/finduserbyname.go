package database

import (
	"database/sql"
	"errors"
	"fmt"
)

func (db *appdbimpl) FindUserByName(name string) (string, error) {
	var identifier string

	// Query: Cerchiamo solo l'identifier, visto che è l'unica cosa che l'API vuole restituire.
	err := db.c.QueryRow("SELECT userId FROM users WHERE userName = ?", name).Scan(&identifier)

	if errors.Is(err, sql.ErrNoRows) {
		// Se l'utente non è trovato, restituiamo il nostro errore specifico per l'API
		return "", ErrUserNotFound
	}
	if err != nil {
		// Qualsiasi altro errore del database
		return "", fmt.Errorf("error querying user by name: %w", err)
	}

	return identifier, nil
}
