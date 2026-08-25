package session

import (
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"agent-ledger/internal/storage"
)

// Session represents an agent session
type Session struct {
	ID        string    `json:"id"`
	Agent     string    `json:"agent,omitempty"`
	Model     string    `json:"model,omitempty"`
	RepoRoot  string    `json:"repo_root"`
	Branch    string    `json:"branch"`
	Head      string    `json:"head"`
	StartTime time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
}

// Manager manages sessions
type Manager struct {
	storage *storage.Storage
}

// NewManager creates a new session manager
func NewManager(st *storage.Storage) *Manager {
	return &Manager{
		storage: st,
	}
}

// Create creates a new session
func (m *Manager) Create(agent, model, repoRoot, branch, head string) (*Session, error) {
	sessionID := uuid.New().String()
	
	session := &Session{
		ID:        sessionID,
		Agent:     agent,
		Model:     model,
		RepoRoot:  repoRoot,
		Branch:    branch,
		Head:      head,
		StartTime: time.Now().UTC(),
	}
	
	// Save session metadata
	sessionPath := fmt.Sprintf("sessions/%s/metadata.json", sessionID)
	if err := m.storage.WriteJSON(sessionPath, session); err != nil {
		return nil, fmt.Errorf("failed to save session metadata: %w", err)
	}
	
	// Update current session
	if err := m.SetCurrent(sessionID); err != nil {
		return nil, fmt.Errorf("failed to set current session: %w", err)
	}
	
	return session, nil
}

// Get retrieves a session by ID
func (m *Manager) Get(sessionID string) (*Session, error) {
	sessionPath := fmt.Sprintf("sessions/%s/metadata.json", sessionID)
	
	var session Session
	if err := m.storage.ReadJSON(sessionPath, &session); err != nil {
		return nil, fmt.Errorf("failed to read session metadata: %w", err)
	}
	
	return &session, nil
}

// GetCurrent retrieves the current session
func (m *Manager) GetCurrent() (*Session, error) {
	var current struct {
		SessionID string `json:"session_id"`
	}
	
	if err := m.storage.ReadJSON("state/current.json", &current); err != nil {
		return nil, fmt.Errorf("failed to read current session: %w", err)
	}
	
	if current.SessionID == "" {
		return nil, fmt.Errorf("no current session")
	}
	
	return m.Get(current.SessionID)
}

// SetCurrent sets the current session
func (m *Manager) SetCurrent(sessionID string) error {
	current := struct {
		SessionID string `json:"session_id"`
	}{
		SessionID: sessionID,
	}
	
	return m.storage.WriteJSON("state/current.json", current)
}

// ClearCurrent clears the current session
func (m *Manager) ClearCurrent() error {
	current := struct {
		SessionID string `json:"session_id"`
	}{
		SessionID: "",
	}
	
	return m.storage.WriteJSON("state/current.json", current)
}

// Stop marks the current session as ended
func (m *Manager) Stop() error {
	session, err := m.GetCurrent()
	if err != nil {
		return err
	}
	
	now := time.Now().UTC()
	session.EndTime = &now
	
	sessionPath := fmt.Sprintf("sessions/%s/metadata.json", session.ID)
	if err := m.storage.WriteJSON(sessionPath, session); err != nil {
		return fmt.Errorf("failed to update session metadata: %w", err)
	}
	
	return m.ClearCurrent()
}

// ListAll lists all sessions
func (m *Manager) ListAll() ([]*Session, error) {
	sessionDirs, err := m.storage.ListDirectories("sessions")
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	
	var sessions []*Session
	for _, dir := range sessionDirs {
		session, err := m.Get(dir)
		if err != nil {
			continue // Skip sessions that can't be read
		}
		sessions = append(sessions, session)
	}
	
	return sessions, nil
}
