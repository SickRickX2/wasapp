package database

import (
	"fmt"
)

func (db *appdbimpl) CreateUser(identifier string, name string) error {
	_, err := db.c.Exec("INSERT INTO Users (userId, userName) VALUES (?, ?)", identifier, name)
	if err != nil {
		// Controlla se l'errore è una violazione di chiave primaria (duplicate key)
		if sqliteErr, ok := err.(interface{ ErrorCode() int }); ok {
			const sqliteConstraintPrimaryKey = 1555 // Codice errore SQLite per violazione di chiave primaria
			if sqliteErr.ErrorCode() == sqliteConstraintPrimaryKey {
				return ErrDuplicateKey
			}
		}
		return fmt.Errorf("error inserting new user: %w", err)
	}
	return nil
}
