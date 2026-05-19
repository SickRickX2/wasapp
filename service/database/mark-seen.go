package database

import (
	"database/sql"
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) MarkAsSeen(convId schemas.ConversationId, messageId schemas.MessageId, readerId schemas.UserId) (schemas.Message, error) {
	var out schemas.Message
	var senderId schemas.UserId
	var convKind string
	err := db.c.QueryRow(
		`SELECT m.senderId, c.kind
		 FROM messages m
		 JOIN conversations c ON c.convId = m.convId
		 WHERE m.messageId = ? AND m.convId = ?`,
		messageId, convId,
	).Scan(&senderId, &convKind)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, errors.New("message not found")
		}
		return out, err
	}

	// se il lettore è il mittente non fa nulla
	if readerId != senderId {
		_, err = db.c.Exec(
			`INSERT OR IGNORE INTO message_reads (messageId, userId, readAt) VALUES (?, ?, CURRENT_TIMESTAMP)`,
			messageId,
			readerId,
		)
		if err != nil {
			return out, err
		}
	}

	// se è privata setta subito seen appena l'altro lo vede
	if convKind == "private" {
		_, err = db.c.Exec(`UPDATE messages SET status = 'seen' WHERE messageId = ? AND status != 'deleted'`, messageId)
		if err != nil {
			return out, err
		}
	}

	// se gruppo tutti tranne mittente devono aver visto
	if convKind == groupType {
		var readersCount int
		err = db.c.QueryRow(`SELECT COUNT(DISTINCT userId) FROM message_reads WHERE messageId = ?`, messageId).Scan(&readersCount)
		if err != nil {
			return out, err
		}

		var expectedReaders int
		err = db.c.QueryRow(
			`SELECT COUNT(*) FROM conversation_participants WHERE convId = ? AND userId != ?`,
			convId,
			senderId,
		).Scan(&expectedReaders)
		if err != nil {
			return out, err
		}

		if readersCount >= expectedReaders {
			_, err = db.c.Exec(`UPDATE messages SET status = 'seen' WHERE messageId = ? AND status != 'deleted'`, messageId)
			if err != nil {
				return out, err
			}
		}
	}
	out, err = db.GetMessage(messageId)
	if err != nil {
		return out, err
	}

	return out, nil
}
