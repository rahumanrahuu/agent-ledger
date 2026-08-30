package collaboration

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Agent represents an agent in the collaboration system
type Agent struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Model    string    `json:"model"`
	Status   string    `json:"status"` // active, idle, working, blocked
	LastSeen time.Time `json:"last_seen"`
	Context  string    `json:"context"`
}

// Result represents a work result shared between agents
type Result struct {
	ID          string      `json:"id"`
	AgentID     string      `json:"agent_id"`
	AgentName   string      `json:"agent_name"`
	Type        string      `json:"type"` // success, failure, insight, error
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Quality     float64     `json:"quality"` // 0.0-1.0
	Timestamp   time.Time   `json:"timestamp"`
	Tags        []string    `json:"tags"`
	Archived    bool        `json:"archived"`
}

// Coordination represents work coordination between agents
type Coordination struct {
	ID            string    `json:"id"`
	ParentAgentID string    `json:"parent_agent_id"`
	ChildAgentID  string    `json:"child_agent_id"`
	Status        string    `json:"status"` // pending, active, completed, failed
	Task          string    `json:"task"`
	Result        *Result   `json:"result,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// Coordinator manages multi-agent collaboration
type Coordinator struct {
	mu            sync.RWMutex
	agents        map[string]*Agent
	results       map[string]*Result
	coordinations map[string]*Coordination
	resultIndex   map[string][]string // agentID -> resultIDs
}

// NewCoordinator creates a new coordinator
func NewCoordinator() *Coordinator {
	return &Coordinator{
		agents:        make(map[string]*Agent),
		results:       make(map[string]*Result),
		coordinations: make(map[string]*Coordination),
		resultIndex:   make(map[string][]string),
	}
}

// RegisterAgent registers an agent in the collaboration system
func (c *Coordinator) RegisterAgent(name, model string) (*Agent, error) {
	if name == "" {
		return nil, fmt.Errorf("agent name is required")
	}

	agent := &Agent{
		ID:       uuid.New().String(),
		Name:     name,
		Model:    model,
		Status:   "idle",
		LastSeen: time.Now(),
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.agents[agent.ID] = agent
	c.resultIndex[agent.ID] = []string{}

	return agent, nil
}

// UpdateAgentStatus updates an agent's status
func (c *Coordinator) UpdateAgentStatus(agentID, status string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	agent, exists := c.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	agent.Status = status
	agent.LastSeen = time.Now()
	return nil
}

// ShareResult shares a result with other agents
func (c *Coordinator) ShareResult(agentID string, title, description string, data interface{}, quality float64) (*Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	agent, exists := c.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	if quality < 0 || quality > 1 {
		quality = 0.5
	}

	result := &Result{
		ID:          uuid.New().String(),
		AgentID:     agentID,
		AgentName:   agent.Name,
		Type:        "success",
		Title:       title,
		Description: description,
		Data:        data,
		Quality:     quality,
		Timestamp:   time.Now(),
	}

	c.results[result.ID] = result
	c.resultIndex[agentID] = append(c.resultIndex[agentID], result.ID)

	return result, nil
}

// ShareFailure shares a failure with other agents
func (c *Coordinator) ShareFailure(agentID string, title, description string, data interface{}) (*Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	agent, exists := c.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	result := &Result{
		ID:          uuid.New().String(),
		AgentID:     agentID,
		AgentName:   agent.Name,
		Type:        "failure",
		Title:       title,
		Description: description,
		Data:        data,
		Quality:     0.0,
		Timestamp:   time.Now(),
	}

	c.results[result.ID] = result
	c.resultIndex[agentID] = append(c.resultIndex[agentID], result.ID)

	return result, nil
}

// GetResults retrieves all results from an agent
func (c *Coordinator) GetResults(agentID string, limit int) []*Result {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resultIDs, exists := c.resultIndex[agentID]
	if !exists {
		return []*Result{}
	}

	var results []*Result
	for _, id := range resultIDs {
		if result, ok := c.results[id]; ok && !result.Archived {
			results = append(results, result)
		}
	}

	// Sort by timestamp descending
	if len(results) > limit && limit > 0 {
		results = results[:limit]
	}

	return results
}

// GetAllResults retrieves all results from all agents with optional filtering
func (c *Coordinator) GetAllResults(limit int) []*Result {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []*Result
	for _, result := range c.results {
		if !result.Archived {
			results = append(results, result)
		}
	}

	if len(results) > limit && limit > 0 {
		results = results[:limit]
	}

	return results
}

// InitiateCoordination initiates work coordination between agents
func (c *Coordinator) InitiateCoordination(parentID, childID, task string) (*Coordination, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.agents[parentID]; !exists {
		return nil, fmt.Errorf("parent agent not found: %s", parentID)
	}
	if _, exists := c.agents[childID]; !exists {
		return nil, fmt.Errorf("child agent not found: %s", childID)
	}

	coord := &Coordination{
		ID:            uuid.New().String(),
		ParentAgentID: parentID,
		ChildAgentID:  childID,
		Status:        "pending",
		Task:          task,
		CreatedAt:     time.Now(),
	}

	c.coordinations[coord.ID] = coord
	return coord, nil
}

// UpdateCoordination updates coordination status
func (c *Coordinator) UpdateCoordination(coordID, status string, result *Result) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	coord, exists := c.coordinations[coordID]
	if !exists {
		return fmt.Errorf("coordination not found: %s", coordID)
	}

	coord.Status = status
	if result != nil {
		coord.Result = result
	}
	if status == "completed" || status == "failed" {
		now := time.Now()
		coord.CompletedAt = &now
	}

	return nil
}

// GetCoordinations retrieves coordinations for an agent
func (c *Coordinator) GetCoordinations(agentID string) []*Coordination {
	c.mu.RLock()
	defer c.mu.Unlock()

	var coords []*Coordination
	for _, coord := range c.coordinations {
		if coord.ParentAgentID == agentID || coord.ChildAgentID == agentID {
			coords = append(coords, coord)
		}
	}

	return coords
}

// GetRecentActivity gets recent activity across all agents
func (c *Coordinator) GetRecentActivity(duration time.Duration, limit int) []*Result {
	c.mu.RLock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-duration)
	var recent []*Result

	for _, result := range c.results {
		if result.Timestamp.After(cutoff) && !result.Archived {
			recent = append(recent, result)
		}
	}

	if len(recent) > limit && limit > 0 {
		recent = recent[:limit]
	}

	return recent
}

// GetAgentStats retrieves statistics for an agent
func (c *Coordinator) GetAgentStats(agentID string) map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	agent, exists := c.agents[agentID]
	if !exists {
		return nil
	}

	resultIDs := c.resultIndex[agentID]
	var successCount, failureCount int
	var avgQuality float64

	for _, id := range resultIDs {
		if result, ok := c.results[id]; ok && !result.Archived {
			if result.Type == "success" {
				successCount++
			} else if result.Type == "failure" {
				failureCount++
			}
			avgQuality += result.Quality
		}
	}

	totalResults := successCount + failureCount
	if totalResults > 0 {
		avgQuality /= float64(totalResults)
	}

	var coordCount int
	for _, coord := range c.coordinations {
		if coord.ParentAgentID == agentID || coord.ChildAgentID == agentID {
			coordCount++
		}
	}

	return map[string]interface{}{
		"agent_id":       agent.ID,
		"agent_name":     agent.Name,
		"status":         agent.Status,
		"total_results":  totalResults,
		"success_count":  successCount,
		"failure_count":  failureCount,
		"avg_quality":    avgQuality,
		"success_rate":   float64(successCount) / float64(totalResults),
		"coordinations":  coordCount,
		"last_seen":      agent.LastSeen,
	}
}

// ListAgents lists all registered agents
func (c *Coordinator) ListAgents() []*Agent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var agents []*Agent
	for _, agent := range c.agents {
		agents = append(agents, agent)
	}

	return agents
}
