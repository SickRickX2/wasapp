package database

import (
	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) SearchUsers(query string) ([]schemas.User, error) {
	var users []schemas.User

	// Es: se query="Ric", cerca "%Ric%"
	searchQuery := "%" + query + "%"

	const sqlQuery = `
		SELECT userId, userName, createdAt, IFNULL(pfpUrl, '') 
		FROM users 
		WHERE userName LIKE ?
	`

	rows, err := db.c.Query(sqlQuery, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u schemas.User
		// riga per riga
		if err := rows.Scan(&u.ID, &u.Name, &u.CreatedAt, &u.PFPURL); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// se non trova nessuno, ritorna array vuoto
	if users == nil {
		users = make([]schemas.User, 0)
	}

	return users, nil
}
