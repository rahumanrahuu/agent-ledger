package api

import (
	"encoding/json"
	"time"

	"agent-ledger/internal/checkpoint"
	"agent-ledger/internal/events"
	"agent-ledger/internal/history"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/session"
	"agent-ledger/internal/storage"
)

// API provides read-only access to Agent Ledger data
type API struct {
	repo          *repository.Repository
	storage       *storage.Storage
	sessionMgr    *session.Manager
	checkpointMgr *checkpoint.Manager
	historyMgr    *history.Manager
	eventsMgr     *events.Manager
	version       string
}

// NewAPI creates a new API instance
func NewAPI(repo *repository.Repository, st *storage.Storage, version string) *API {
	sessionMgr := session.NewManager(st)
	checkpointMgr := checkpoint.NewManager(st)
	historyMgr := history.NewManager(sessionMgr, checkpointMgr, st)
	eventsMgr := events.NewManager(st)

	return &API{
		repo:          repo,
		storage:       st,
		sessionMgr:    sessionMgr,
		checkpointMgr: checkpointMgr,
		historyMgr:    historyMgr,
		eventsMgr:     eventsMgr,
		version:       version,
	}
}

// OverviewResponse contains project overview data
type OverviewResponse struct {
	ProjectName       string    `json:"project_name"`
	RepositoryRoot    string    `json:"repository_root"`
	CurrentBranch     string    `json:"current_branch"`
	CurrentCommit     string    `json:"current_commit"`
	Version           string    `json:"version"`
	SessionCount      int       `json:"session_count"`
	DecisionCount     int       `json:"decision_count"`
	DiscoveryCount    int       `json:"discovery_count"`
	CheckpointCount   int       `json:"checkpoint_count"`
	LastActivityTime  *time.Time `json:"last_activity_time,omitempty"`
}

// GetOverview returns project overview data
func (a *API) GetOverview() (*OverviewResponse, error) {
	sessions, err := a.historyMgr.GetAllSessions("", "")
	if err != nil {
		sessions = []*session.Session{}
	}

	decisions, _ := a.storage.ListFiles("decisions")
	discoveries, _ := a.storage.ListFiles("discoveries")
	checkpoints, _ := a.storage.ListFiles("checkpoints")

	projectName := "Agent Ledger Project"
	if content, err := a.storage.ReadMarkdown("project.md"); err == nil && len(content) > 0 {
		// Try to extract project name from project.md
		lines := make([]rune, 0, len(content))
		for _, r := range content {
			lines = append(lines, r)
		}
		if len(lines) > 0 {
			projectName = "Project"
		}
	}

	return &OverviewResponse{
		ProjectName:     projectName,
		RepositoryRoot:  a.repo.Root,
		CurrentBranch:   a.repo.Branch,
		CurrentCommit:   a.repo.Head,
		Version:         a.version,
		SessionCount:    len(sessions),
		DecisionCount:   len(decisions),
		DiscoveryCount:  len(discoveries),
		CheckpointCount: len(checkpoints),
	}, nil
}

// SessionListResponse contains a list of sessions
type SessionListResponse struct {
	Sessions []*SessionInfo `json:"sessions"`
}

// SessionInfo contains basic session information
type SessionInfo struct {
	ID        string     `json:"id"`
	Agent     string     `json:"agent,omitempty"`
	Model     string     `json:"model,omitempty"`
	Branch    string     `json:"branch"`
	Head      string     `json:"head"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Status    string     `json:"status"`
}

// GetSessions returns all sessions
func (a *API) GetSessions() (*SessionListResponse, error) {
	sessionList, err := a.historyMgr.GetAllSessions("", "")
	if err != nil {
		return &SessionListResponse{Sessions: []*SessionInfo{}}, nil
	}

	sessions := make([]*SessionInfo, len(sessionList))
	for i, s := range sessionList {
		status := "ended"
		if s.EndTime == nil {
			status = "active"
		}

		sessions[i] = &SessionInfo{
			ID:        s.ID,
			Agent:     s.Agent,
			Model:     s.Model,
			Branch:    s.Branch,
			Head:      s.Head,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			Status:    status,
		}
	}

	return &SessionListResponse{Sessions: sessions}, nil
}

// SessionDetailResponse contains detailed session information
type SessionDetailResponse struct {
	Session      *SessionInfo `json:"session"`
	Checkpoints  int          `json:"checkpoint_count"`
	Decisions    int          `json:"decision_count"`
	Discoveries  int          `json:"discovery_count"`
	Failures     int          `json:"failure_count"`
	Constraints  int          `json:"constraint_count"`
	HasHandoff   bool         `json:"has_handoff"`
}

// GetSessionDetail returns detailed information about a specific session
func (a *API) GetSessionDetail(sessionID string) (*SessionDetailResponse, error) {
	sess, err := a.sessionMgr.Get(sessionID)
	if err != nil {
		return nil, err
	}

	status := "ended"
	if sess.EndTime == nil {
		status = "active"
	}

	sessionInfo := &SessionInfo{
		ID:        sess.ID,
		Agent:     sess.Agent,
		Model:     sess.Model,
		Branch:    sess.Branch,
		Head:      sess.Head,
		StartTime: sess.StartTime,
		EndTime:   sess.EndTime,
		Status:    status,
	}

	// Get session history to count artifacts
	history, err := a.historyMgr.GetSessionHistory(sessionID)
	if err != nil {
		return &SessionDetailResponse{
			Session:     sessionInfo,
			Checkpoints: 0,
			Decisions:   0,
			Discoveries: 0,
			Failures:    0,
			Constraints: 0,
			HasHandoff:  false,
		}, nil
	}

	return &SessionDetailResponse{
		Session:     sessionInfo,
		Checkpoints: len(history.Checkpoints),
		Decisions:   len(history.Decisions),
		Discoveries: len(history.Discoveries),
		Failures:    len(history.Failures),
		Constraints: len(history.Constraints),
		HasHandoff:  history.Handoff != "",
	}, nil
}

// EventListResponse contains a list of events
type EventListResponse struct {
	Events []json.RawMessage `json:"events"`
}

// GetEvents returns all events (decisions, discoveries, checkpoints, etc.) in chronological order
func (a *API) GetEvents(filterType string) (*EventListResponse, error) {
	var eventList []json.RawMessage

	// Collect all events with timestamps for sorting
	type timedEvent struct {
		timestamp time.Time
		data      json.RawMessage
	}

	var events []timedEvent

	// Get decisions
	if filterType == "" || filterType == "decision" {
		decisionFiles, _ := a.storage.ListFiles("decisions")
		for _, file := range decisionFiles {
			path := "decisions/" + file
			content, _ := a.storage.ReadMarkdown(path)
			if len(content) > 0 {
				// Parse as decision - simplified for now
				events = append(events, timedEvent{
					timestamp: time.Now(),
					data:      json.RawMessage(content),
				})
			}
		}
	}

	// For now, return empty list - will be expanded
	return &EventListResponse{Events: eventList}, nil
}

// GraphResponse contains knowledge graph data
type GraphResponse struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node represents a node in the knowledge graph
type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"` // "session", "decision", "discovery", "checkpoint"
	Data  any    `json:"data,omitempty"`
}

// Edge represents an edge in the knowledge graph
type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"` // "contains", "relates_to", etc.
}

// GetGraph returns the knowledge graph
func (a *API) GetGraph() (*GraphResponse, error) {
	var nodes []Node
	var edges []Edge

	// Get sessions
	sessions, _ := a.historyMgr.GetAllSessions("", "")
	for _, s := range sessions {
		nodes = append(nodes, Node{
			ID:    s.ID,
			Label: s.Agent,
			Type:  "session",
			Data:  s,
		})
	}

	// TODO: Add decisions, discoveries, checkpoints, and relationships

	return &GraphResponse{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

// SearchResponse contains search results
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "session", "decision", etc.
	Title    string `json:"title"`
	Excerpt  string `json:"excerpt"`
	Path     string `json:"path"`
}

// Search searches across all entities
func (a *API) Search(query string) (*SearchResponse, error) {
	// TODO: Implement search
	return &SearchResponse{Results: []SearchResult{}}, nil
}
