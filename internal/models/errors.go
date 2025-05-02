package models

import "errors"

var ErrNoRecord = errors.New("models: no matching record found")
var ErrSessionNotFound = errors.New("session: no session found")
