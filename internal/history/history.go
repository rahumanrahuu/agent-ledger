package history

import (
	"fmt"
	"time"
	
	"agent-ledger/internal/checkpoint"
	"agent-ledger/internal/session"
	"agent-ledger/internal/storage"
)

// Manager manages history
type Manager struct {
	sessionManager    *session.Manager
	checkpointManager *checkpoint.Manager
	storage           *storage.Storage
}

// NewManager creates a new history manager
func NewManager(sessionManager *session.Manager, checkpointManager *checkpoint.Manager, st *storage.Storage) *Manager {
	return &Manager{
		sessionManager:    sessionManager,
		checkpointManager: checkpointManager,
		storage:           st,
	}
}

// SessionHistory represents the history of a session
type SessionHistory struct {
	Session     *session.Session
	Checkpoints []checkpoint.Checkpoint
	Decisions   []string
	Discoveries []string
	Failures    []string
	Constraints []string
	Handoff     string
}

// GetSessionHistory gets the complete history of a session
func (m *Manager) GetSessionHistory(sessionID string) (*SessionHistory, error) {
	sess, err := m.sessionManager.Get(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	checkpoints, err := m.checkpointManager.List(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoints: %w", err)
	}
	
	// Get events
	decisions, err := m.storage.ListFiles("decisions")
	if err != nil {
		decisions = []string{}
	}
	
	discoveries, err := m.storage.ListFiles("discoveries")
	if err != nil {
		discoveries = []string{}
	}
	
	failures, err := m.storage.ListFiles("failures")
	if err != nil {
		failures = []string{}
	}
	
	constraints, err := m.storage.ListFiles("constraints")
	if err != nil {
		constraints = []string{}
	}
	
	// Get handoff
	var handoff string
	handoffPath := fmt.Sprintf("sessions/%s/handoff.md", sessionID)
	if m.storage.FileExists(handoffPath) {
		handoffContent, err := m.storage.ReadMarkdown(handoffPath)
		if err == nil {
			handoff = handoffContent
		}
	}
	
	return &SessionHistory{
		Session:     sess,
		Checkpoints: checkpoints,
		Decisions:   decisions,
		Discoveries: discoveries,
		Failures:    failures,
		Constraints: constraints,
		Handoff:     handoff,
	}, nil
}

// GetAllSessions gets all sessions with optional filtering
func (m *Manager) GetAllSessions(agent string, model string) ([]*session.Session, error) {
	sessions, err := m.sessionManager.ListAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	
	// Filter by agent if specified
	if agent != "" {
		var filtered []*session.Session
		for _, sess := range sessions {
			if sess.Agent == agent {
				filtered = append(filtered, sess)
			}
		}
		sessions = filtered
	}
	
	// Filter by model if specified
	if model != "" {
		var filtered []*session.Session
		for _, sess := range sessions {
			if sess.Model == model {
				filtered = append(filtered, sess)
			}
		}
		sessions = filtered
	}
	
	return sessions, nil
}

// GetRecentSessions gets recent sessions within a time range
func (m *Manager) GetRecentSessions(since time.Time) ([]*session.Session, error) {
	sessions, err := m.sessionManager.ListAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	
	var recent []*session.Session
	for _, sess := range sessions {
		if sess.StartTime.After(since) || sess.StartTime.Equal(since) {
			recent = append(recent, sess)
		}
	}
	
	return recent, nil
}

// GetActiveSession gets the currently active session
func (m *Manager) GetActiveSession() (*session.Session, error) {
	return m.sessionManager.GetCurrent()
}
