package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) SaveMedia(media schemas.Media, mediaId string) error {
	const query = `
		INSERT INTO media (mediaId, url, filename, mimeType, size, createdAt)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	// Nota: L'URL che salviamo è relativo, es: "/images/xyz.jpg"
	_, err := db.c.Exec(query, mediaId, media.URL, media.Filename, media.MimeType, media.Size)
	return err
}
