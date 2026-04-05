package database

import "github.com/SickRickX2/wasapp/service/api/schemas"

func (db *appdbimpl) CountUsersByNameExcludingUser(name string, excludeUserID schemas.UserId) (int, error) {
	const query = `SELECT COUNT(*) FROM users WHERE userName = ? AND userId <> ?`
	var count int
	err := db.c.QueryRow(query, name, excludeUserID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *appdbimpl) SetUserName(id schemas.UserId, newName string) error {
	const query = `UPDATE users SET userName = ? WHERE userId = ?`
	_, err := db.c.Exec(query, newName, id)
	return err
}
