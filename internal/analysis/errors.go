package analysis

import "errors"

var (
	ErrSessionNotCompleted = errors.New("session not completed, analysis only available for completed sessions")
	ErrSessionNotFound     = errors.New("session not found")
)
