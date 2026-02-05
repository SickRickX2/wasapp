package database

import (
	"database/sql" // <--- Serve per gestire i NULL del database (sql.NullString)

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) GetConversationMessages(convId schemas.ConversationId, limit int, beforeId string) ([]schemas.Message, error) {
	var messages []schemas.Message

	// 1. MODIFICA QUI: Ho aggiunto 'mediaId' alla SELECT
	query := `SELECT messageId, senderId, text, mediaId, sentAt, kind, status FROM messages WHERE convId = ?`
	args := []interface{}{convId}

	// prendiamo i messaggi prima di beforeid
	if beforeId != "" {
		query += ` AND sentAt < (SELECT sentAt FROM messages WHERE messageId = ?)`
		args = append(args, beforeId)
	}

	// ordiniamo
	query += ` ORDER BY sentAt DESC LIMIT ?`
	args = append(args, limit)

	rows, err := db.c.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m schemas.Message

		// 2. MODIFICA QUI: Variabile per gestire il fatto che mediaId può essere NULL
		var mediaIdSql sql.NullString

		// 3. MODIFICA QUI: Ho aggiunto &mediaIdSql allo Scan (nell'ordine giusto della query sopra)
		if err := rows.Scan(&m.ID, &m.Sender, &m.Text, &mediaIdSql, &m.Time, &m.Kind, &m.Status); err != nil {
			return nil, err
		}

		// Se c'è un mediaId nel DB, lo copiamo nella struct
		if mediaIdSql.Valid {
			m.MediaId = mediaIdSql.String
		}

		// 4. MODIFICA QUI: INIZIO LOGICA REAZIONI
		// Invece di lasciarle vuote, chiediamo al DB chi ha reagito a QUESTO messaggio
		// Nota: Assicurati che la tabella si chiami 'message_reactions' come abbiamo detto prima
		reacRows, err := db.c.Query(`SELECT userId, emoji FROM message_reactions WHERE messageId = ?`, m.ID)
		if err != nil {
			return nil, err // O logga l'errore e continua, a tua scelta
		}

		m.Reactions = []schemas.Reaction{} // Inizializza array vuoto

		for reacRows.Next() {
			var r schemas.Reaction
			// Leggiamo userId ed emoji
			if err := reacRows.Scan(&r.UserId, &r.Emoji); err != nil {
				reacRows.Close()
				return nil, err
			}
			m.Reactions = append(m.Reactions, r)
		}
		reacRows.Close() // Chiudiamo subito il cursore secondario
		// FINE LOGICA REAZIONI ----------------------

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// non deve restituire nil
	if messages == nil {
		messages = make([]schemas.Message, 0)
	}

	return messages, nil
}
