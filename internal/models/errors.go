package models

import "errors"

var (
	ErrNoRecord = errors.New("models: no matching record found")
	ErrSessionNotFound = errors.New("session: no session found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail = errors.New("models: duplicate email")
)

