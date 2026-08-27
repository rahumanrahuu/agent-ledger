package memory

import "errors"

var (
	ErrSessionNotFound   = errors.New("session not found")
	ErrMemoryNotFound    = errors.New("memory not found")
	ErrAgentNotFound     = errors.New("agent not found")
	ErrConflictDetected  = errors.New("memory conflict detected")
	ErrDatabaseError     = errors.New("database error")
	ErrInvalidQuery      = errors.New("invalid query")
)
