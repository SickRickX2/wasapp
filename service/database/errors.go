package database

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrConversationNotFound = errors.New("conversation not found")
var ErrMessageNotFound = errors.New("message not found")
var ErrSessionNotFound = errors.New("session not found")
var ErrGroupNotFound = errors.New("group not found")
var ErrParticipantNotFound = errors.New("participant not found")
var ErrMediaNotFound = errors.New("media not found")
var ErrDuplicateKey = errors.New("duplicate key violation")
var ErrAlreadyExists = errors.New("resource already exists")
var ErrForbidden = errors.New("forbidden")
