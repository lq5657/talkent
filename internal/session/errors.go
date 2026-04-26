package session

import "errors"

var (
	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionCompleted  = errors.New("session already completed")
)
