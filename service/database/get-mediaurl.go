package database

import (
	"database/sql"
	"errors"
)

// GetMediaUrl recupera l'URL pubblico di un media partendo dal suo ID
func (db *appdbimpl) GetMediaUrl(mediaId string) (string, error) {
	var url string

	query := `SELECT url FROM media WHERE mediaId = ?`

	err := db.c.QueryRow(query, mediaId).Scan(&url)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("media not found")
		}
		return "", err
	}

	return url, nil
}
