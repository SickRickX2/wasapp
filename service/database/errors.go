package database

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrDuplicateKey = errors.New("duplicate key violation")
