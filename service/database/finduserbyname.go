package database

import (
	"database/sql"
	"errors"

	"github.com/SickRickX2/wasapp/service/api/schemas"
)

func (db *appdbimpl) FindUserByName(name string) (schemas.User, error) {
	var u schemas.User
	const query = `
		SELECT userId, userName, createdAt, IFNULL(pfpUrl, '') 
		FROM users 
		WHERE userName = ?
	`
	// NOTA: Scan su &u.ID e &u.Name
	err := db.c.QueryRow(query, name).Scan(&u.ID, &u.Name, &u.CreatedAt, &u.PFPURL)

	if errors.Is(err, sql.ErrNoRows) {
		return u, errors.New("user not found")
	}
	if err != nil {
		return u, err
	}
	return u, nil
}

func (db *appdbimpl) GetUserById(id schemas.UserId) (schemas.User, error) {
	var u schemas.User
	const query = `
		SELECT userId, userName, createdAt, IFNULL(pfpUrl, '') 
		FROM users 
		WHERE userId = ?
	`
	err := db.c.QueryRow(query, id).Scan(&u.ID, &u.Name, &u.CreatedAt, &u.PFPURL)
	if errors.Is(err, sql.ErrNoRows) {
		return u, errors.New("user not found")
	}
	return u, err
}
