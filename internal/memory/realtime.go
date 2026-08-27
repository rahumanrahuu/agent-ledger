package memory

import (
	"encoding/json"
	"sync"
	"time"
)

// CollaborationSession represents a real-time collaboration session
type CollaborationSession struct {
	ID         string
	CreatedAt  time.Time
	LastUpdate time.Time
	Participants map[string]*Participant
	ActiveMemories map[string]*Memory
	mu         sync.RWMutex
}

// Participant represents a participant in collaboration
type Participant struct {
	ID           string
	Name         string
	JoinedAt     time.Time
	LastActivity time.Time
	Cursor       CursorPosition
	Status       string // active, idle, away
}

// CursorPosition tracks cursor position for live editing
type CursorPosition struct {
	MemoryID string
	Line     int
	Column   int
	Selection string
}

// SyncMessage represents a message sent over WebSocket
type SyncMessage struct {
	Type        string                 `json:"type"` // memory_update, cursor_move, presence, conflict
	SessionID   string                 `json:"session_id"`
	ParticipantID string                `json:"participant_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	MemoryID    string                 `json:"memory_id,omitempty"`
}

// CollaborationHub manages real-time collaboration
type CollaborationHub struct {
	sessions map[string]*CollaborationSession
	clients  map[string]chan SyncMessage
	mu       sync.RWMutex
}

// NewCollaborationHub creates a new collaboration hub
func NewCollaborationHub() *CollaborationHub {
	return &CollaborationHub{
		sessions: make(map[string]*CollaborationSession),
		clients:  make(map[string]chan SyncMessage),
	}
}

// CreateSession creates a new collaboration session
func (ch *CollaborationHub) CreateSession(sessionID string) *CollaborationSession {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	session := &CollaborationSession{
		ID:         sessionID,
		CreatedAt:  time.Now(),
		LastUpdate: time.Now(),
		Participants: make(map[string]*Participant),
		ActiveMemories: make(map[string]*Memory),
	}

	ch.sessions[sessionID] = session
	return session
}

// JoinSession adds a participant to a session
func (ch *CollaborationHub) JoinSession(sessionID string, participant Participant) error {
	ch.mu.RLock()
	session, ok := ch.sessions[sessionID]
	ch.mu.RUnlock()

	if !ok {
		return ErrSessionNotFound
	}

	session.mu.Lock()
	participant.JoinedAt = time.Now()
	participant.LastActivity = time.Now()
	participant.Status = "active"
	session.Participants[participant.ID] = &participant
	session.LastUpdate = time.Now()
	session.mu.Unlock()

	// Broadcast presence
	ch.Broadcast(SyncMessage{
		Type:        "presence",
		SessionID:   sessionID,
		ParticipantID: participant.ID,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"action": "joined",
			"name":   participant.Name,
		},
	})

	return nil
}

// LeaveSession removes a participant from a session
func (ch *CollaborationHub) LeaveSession(sessionID string, participantID string) error {
	ch.mu.RLock()
	session, ok := ch.sessions[sessionID]
	ch.mu.RUnlock()

	if !ok {
		return ErrSessionNotFound
	}

	session.mu.Lock()
	delete(session.Participants, participantID)
	session.LastUpdate = time.Now()
	session.mu.Unlock()

	// Broadcast presence
	ch.Broadcast(SyncMessage{
		Type:        "presence",
		SessionID:   sessionID,
		ParticipantID: participantID,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"action": "left",
		},
	})

	return nil
}

// UpdateMemory updates a memory in a session
func (ch *CollaborationHub) UpdateMemory(sessionID string, participantID string, memory Memory) error {
	ch.mu.RLock()
	session, ok := ch.sessions[sessionID]
	ch.mu.RUnlock()

	if !ok {
		return ErrSessionNotFound
	}

	session.mu.Lock()
	session.ActiveMemories[memory.ID] = &memory
	session.LastUpdate = time.Now()
	session.mu.Unlock()

	// Update participant activity
	ch.UpdateParticipantActivity(sessionID, participantID)

	// Broadcast update
	data, _ := json.Marshal(memory)
	ch.Broadcast(SyncMessage{
		Type:        "memory_update",
		SessionID:   sessionID,
		ParticipantID: participantID,
		Timestamp:   time.Now(),
		MemoryID:    memory.ID,
		Data: map[string]interface{}{
			"memory": json.RawMessage(data),
		},
	})

	return nil
}

// UpdateCursor updates participant cursor position
func (ch *CollaborationHub) UpdateCursor(sessionID string, participantID string, position CursorPosition) error {
	ch.mu.RLock()
	session, ok := ch.sessions[sessionID]
	ch.mu.RUnlock()

	if !ok {
		return ErrSessionNotFound
	}

	session.mu.Lock()
	if participant, ok := session.Participants[participantID]; ok {
		participant.Cursor = position
		participant.LastActivity = time.Now()
	}
	session.LastUpdate = time.Now()
	session.mu.Unlock()

	// Broadcast cursor
	ch.Broadcast(SyncMessage{
		Type:        "cursor_move",
		SessionID:   sessionID,
		ParticipantID: participantID,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"memory_id": position.MemoryID,
			"line":      position.Line,
			"column":    position.Column,
		},
	})

	return nil
}

// UpdateParticipantActivity updates last activity time
func (ch *CollaborationHub) UpdateParticipantActivity(sessionID string, participantID string) {
	ch.mu.RLock()
	session, ok := ch.sessions[sessionID]
	ch.mu.RUnlock()

	if !ok {
		return
	}

	session.mu.Lock()
	if participant, ok := session.Participants[participantID]; ok {
		participant.LastActivity = time.Now()
	}
	session.mu.Unlock()
}

// GetActiveParticipants gets all active participants
func (ch *CollaborationHub) GetActiveParticipants(sessionID string) []*Participant {
	ch.mu.RLock()
	session, ok := ch.sessions[sessionID]
	ch.mu.RUnlock()

	if !ok {
		return nil
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	var active []*Participant
	now := time.Now()

	for _, p := range session.Participants {
		if now.Sub(p.LastActivity) < 5*time.Minute {
			active = append(active, p)
		}
	}

	return active
}

// GetSessionMemories gets all active memories in a session
func (ch *CollaborationHub) GetSessionMemories(sessionID string) []*Memory {
	ch.mu.RLock()
	session, ok := ch.sessions[sessionID]
	ch.mu.RUnlock()

	if !ok {
		return nil
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	var memories []*Memory
	for _, m := range session.ActiveMemories {
		memories = append(memories, m)
	}

	return memories
}

// Broadcast sends a message to all connected clients
func (ch *CollaborationHub) Broadcast(message SyncMessage) {
	ch.mu.RLock()
	clients := ch.clients
	ch.mu.RUnlock()

	for _, clientChan := range clients {
		select {
		case clientChan <- message:
		default:
			// Client buffer full, skip
		}
	}
}

// RegisterClient registers a new client
func (ch *CollaborationHub) RegisterClient(clientID string) <-chan SyncMessage {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	messageChan := make(chan SyncMessage, 100)
	ch.clients[clientID] = messageChan

	return messageChan
}

// UnregisterClient unregisters a client
func (ch *CollaborationHub) UnregisterClient(clientID string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if messageChan, ok := ch.clients[clientID]; ok {
		close(messageChan)
		delete(ch.clients, clientID)
	}
}

// ListSessions lists all active sessions
type SessionInfo struct {
	ID           string
	CreatedAt    time.Time
	LastUpdate   time.Time
	ParticipantCount int
	MemoryCount  int
}

// GetSessions returns info about all sessions
func (ch *CollaborationHub) GetSessions() []SessionInfo {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	var sessions []SessionInfo

	for _, session := range ch.sessions {
		session.mu.RLock()
		info := SessionInfo{
			ID:           session.ID,
			CreatedAt:    session.CreatedAt,
			LastUpdate:   session.LastUpdate,
			ParticipantCount: len(session.Participants),
			MemoryCount:  len(session.ActiveMemories),
		}
		session.mu.RUnlock()

		sessions = append(sessions, info)
	}

	return sessions
}
