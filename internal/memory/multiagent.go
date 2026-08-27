package memory

import (
	"fmt"
	"sync"
	"time"
)

// Agent represents an agent in the system
type Agent struct {
	ID            string
	Name          string
	Type          string
	ConnectedAt   time.Time
	LastHeartbeat time.Time
	Status        string // active, idle, error
}

// SharedMemoryHub manages shared memory between agents
type SharedMemoryHub struct {
	agents     map[string]*Agent
	memories   map[string]*Memory
	locks      map[string]*sync.RWMutex
	conflicts  []MemoryConflict
	mu         sync.RWMutex
	eventChan  chan MemoryEvent
}

// MemoryConflict represents a memory conflict
type MemoryConflict struct {
	MemoryID    string
	Timestamp   time.Time
	Agent1      string
	Agent2      string
	Version1    Memory
	Version2    Memory
	Resolution  string
	ResolvedAt  time.Time
	ResolvedBy  string
}

// MemoryEvent represents an event in shared memory
type MemoryEvent struct {
	Type      string                 `json:"type"` // create, update, delete, conflict
	MemoryID  string                 `json:"memory_id"`
	AgentID   string                 `json:"agent_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewSharedMemoryHub creates a new shared memory hub
func NewSharedMemoryHub() *SharedMemoryHub {
	return &SharedMemoryHub{
		agents:    make(map[string]*Agent),
		memories:  make(map[string]*Memory),
		locks:     make(map[string]*sync.RWMutex),
		conflicts: []MemoryConflict{},
		eventChan: make(chan MemoryEvent, 100),
	}
}

// RegisterAgent registers an agent
func (h *SharedMemoryHub) RegisterAgent(agent Agent) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	agent.ConnectedAt = time.Now()
	agent.LastHeartbeat = time.Now()
	agent.Status = "active"

	h.agents[agent.ID] = &agent

	h.eventChan <- MemoryEvent{
		Type:      "agent_connected",
		AgentID:   agent.ID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"agent_name": agent.Name,
			"agent_type": agent.Type,
		},
	}

	return nil
}

// UnregisterAgent unregisters an agent
func (h *SharedMemoryHub) UnregisterAgent(agentID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.agents, agentID)

	h.eventChan <- MemoryEvent{
		Type:      "agent_disconnected",
		AgentID:   agentID,
		Timestamp: time.Now(),
	}

	return nil
}

// AgentRetrieve retrieves memory visible to an agent
func (h *SharedMemoryHub) AgentRetrieve(agentID string, query string) ([]*Memory, error) {
	h.mu.RLock()
	agent, exists := h.agents[agentID]
	h.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent %s not registered", agentID)
	}

	// Update heartbeat
	h.mu.Lock()
	agent.LastHeartbeat = time.Now()
	h.mu.Unlock()

	var results []*Memory

	h.mu.RLock()
	for _, mem := range h.memories {
		// Check access rights (simplified - full impl would check ACL)
		if mem.Type != "" {
			results = append(results, mem)
		}
	}
	h.mu.RUnlock()

	return results, nil
}

// AgentWrite writes memory from an agent
func (h *SharedMemoryHub) AgentWrite(agentID string, memory Memory) (string, error) {
	h.mu.Lock()
	agent, exists := h.agents[agentID]
	if !exists {
		h.mu.Unlock()
		return "", fmt.Errorf("agent %s not registered", agentID)
	}

	agent.LastHeartbeat = time.Now()

	// Check for conflicts
	if existing, ok := h.memories[memory.ID]; ok {
		if existing.UpdatedAt.After(memory.CreatedAt) {
			// Conflict detected
			conflict := MemoryConflict{
				MemoryID:   memory.ID,
				Timestamp:  time.Now(),
				Agent1:     existing.SessionID,
				Agent2:     agentID,
				Version1:   *existing,
				Version2:   memory,
				Resolution: "pending",
			}
			h.conflicts = append(h.conflicts, conflict)
			h.mu.Unlock()

			h.eventChan <- MemoryEvent{
				Type:      "conflict",
				MemoryID:  memory.ID,
				AgentID:   agentID,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"conflict_id": memory.ID,
				},
			}

			return "", fmt.Errorf("conflict detected for memory %s", memory.ID)
		}
	}

	memory.SessionID = agentID
	memory.UpdatedAt = time.Now()

	h.memories[memory.ID] = &memory
	h.mu.Unlock()

	h.eventChan <- MemoryEvent{
		Type:      "memory_write",
		MemoryID:  memory.ID,
		AgentID:   agentID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"title": memory.Title,
			"type":  memory.Type,
		},
	}

	return memory.ID, nil
}

// ListActiveAgents lists all active agents
func (h *SharedMemoryHub) ListActiveAgents() []*Agent {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var active []*Agent
	now := time.Now()

	for _, agent := range h.agents {
		// Agent is active if heartbeat within last 5 minutes
		if now.Sub(agent.LastHeartbeat) < 5*time.Minute {
			active = append(active, agent)
		}
	}

	return active
}

// GetConflicts gets all conflicts
func (h *SharedMemoryHub) GetConflicts() []MemoryConflict {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.conflicts
}

// ResolveConflict resolves a conflict
func (h *SharedMemoryHub) ResolveConflict(memoryID string, resolution string, resolvedBy string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, conflict := range h.conflicts {
		if conflict.MemoryID == memoryID && conflict.Resolution == "pending" {
			h.conflicts[i].Resolution = resolution
			h.conflicts[i].ResolvedAt = time.Now()
			h.conflicts[i].ResolvedBy = resolvedBy

			// Apply resolution
			if resolution == "keep_v1" {
				// Keep existing version
			} else if resolution == "use_v2" {
				// Use new version
				h.memories[memoryID] = &conflict.Version2
			}

			h.eventChan <- MemoryEvent{
				Type:      "conflict_resolved",
				MemoryID:  memoryID,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"resolution": resolution,
					"resolved_by": resolvedBy,
				},
			}

			return nil
		}
	}

	return fmt.Errorf("no pending conflict found for memory %s", memoryID)
}

// GetEventStream returns the event channel
func (h *SharedMemoryHub) GetEventStream() <-chan MemoryEvent {
	return h.eventChan
}

// BroadcastEvent sends an event to all agents
func (h *SharedMemoryHub) BroadcastEvent(event MemoryEvent) {
	h.eventChan <- event
}

// Stats returns hub statistics
type HubStats struct {
	ActiveAgents    int
	TotalAgents     int
	SharedMemories  int
	PendingConflicts int
	ResolvedConflicts int
}

// GetStats returns hub statistics
func (h *SharedMemoryHub) GetStats() HubStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := HubStats{
		TotalAgents:   len(h.agents),
		SharedMemories: len(h.memories),
	}

	now := time.Now()
	for _, agent := range h.agents {
		if now.Sub(agent.LastHeartbeat) < 5*time.Minute {
			stats.ActiveAgents++
		}
	}

	for _, conflict := range h.conflicts {
		if conflict.Resolution == "pending" {
			stats.PendingConflicts++
		} else {
			stats.ResolvedConflicts++
		}
	}

	return stats
}
