package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) SaveMedia(media schemas.Media, mediaId string) error {
	const query = `
		INSERT INTO media (mediaId, url, filename, mimeType, size, createdAt)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	// salva il media nel database
	_, err := db.c.Exec(query, mediaId, media.URL, media.Filename, media.MimeType, media.Size)
	return err
}
