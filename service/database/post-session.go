package database

import "fmt"

func (db *appdbimpl) PostSession(identifier string) error {
	res, err := db.c.Exec("INSERT INTO users (identifier) VALUES (?)", identifier)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// Se la lettura delle righe affette fallisce, restituisci l'errore.
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("fallimento nell'inserimento della sessione: 0 righe affette")
	}

	return nil

}
