package api

import (
	"fmt"
	"strings"
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
	ProjectName       string     `json:"project_name"`
	RepositoryRoot    string     `json:"repository_root"`
	CurrentBranch     string     `json:"current_branch"`
	CurrentCommit     string     `json:"current_commit"`
	Version           string     `json:"version"`
	SessionCount      int        `json:"session_count"`
	DecisionCount     int        `json:"decision_count"`
	DiscoveryCount    int        `json:"discovery_count"`
	CheckpointCount   int        `json:"checkpoint_count"`
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
		projectName = "Project"
	}

	var lastActivityTime *time.Time
	if len(sessions) > 0 && sessions[len(sessions)-1].EndTime != nil {
		lastActivityTime = sessions[len(sessions)-1].EndTime
	} else if len(sessions) > 0 {
		lastActivityTime = &sessions[len(sessions)-1].StartTime
	}

	return &OverviewResponse{
		ProjectName:      projectName,
		RepositoryRoot:   a.repo.Root,
		CurrentBranch:    a.repo.Branch,
		CurrentCommit:    a.repo.Head,
		Version:          a.version,
		SessionCount:     len(sessions),
		DecisionCount:    len(decisions),
		DiscoveryCount:   len(discoveries),
		CheckpointCount:  len(checkpoints),
		LastActivityTime: lastActivityTime,
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

// EventItem represents a single event in the timeline
type EventItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "decision", "discovery", "failure", "constraint"
	Title     string    `json:"title"`
	Content   string    `json:"content,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

// EventListResponse contains a list of events
type EventListResponse struct {
	Events []*EventItem `json:"events"`
}

// GetEvents returns all events in chronological order
func (a *API) GetEvents(filterType string) (*EventListResponse, error) {
	var events []*EventItem

	// Get decisions
	if filterType == "" || filterType == "decision" {
		decisionFiles, _ := a.storage.ListFiles("decisions")
		for _, file := range decisionFiles {
			path := "decisions/" + file
			content, _ := a.storage.ReadMarkdown(path)
			if len(content) > 0 {
				title := strings.TrimSuffix(file, ".md")
				id := extractIDFromFilename(file)
				events = append(events, &EventItem{
					ID:        id,
					Type:      "decision",
					Title:     title,
					Content:   content,
					Timestamp: extractTimestampFromContent(content),
					Path:      path,
				})
			}
		}
	}

	// Get discoveries
	if filterType == "" || filterType == "discovery" {
		discoveryFiles, _ := a.storage.ListFiles("discoveries")
		for _, file := range discoveryFiles {
			path := "discoveries/" + file
			content, _ := a.storage.ReadMarkdown(path)
			if len(content) > 0 {
				title := strings.TrimSuffix(file, ".md")
				id := extractIDFromFilename(file)
				events = append(events, &EventItem{
					ID:        id,
					Type:      "discovery",
					Title:     title,
					Content:   content,
					Timestamp: extractTimestampFromContent(content),
					Path:      path,
				})
			}
		}
	}

	// Get failures
	if filterType == "" || filterType == "failure" {
		failureFiles, _ := a.storage.ListFiles("failures")
		for _, file := range failureFiles {
			path := "failures/" + file
			content, _ := a.storage.ReadMarkdown(path)
			if len(content) > 0 {
				title := strings.TrimSuffix(file, ".md")
				id := extractIDFromFilename(file)
				events = append(events, &EventItem{
					ID:        id,
					Type:      "failure",
					Title:     title,
					Content:   content,
					Timestamp: extractTimestampFromContent(content),
					Path:      path,
				})
			}
		}
	}

	// Get constraints
	if filterType == "" || filterType == "constraint" {
		constraintFiles, _ := a.storage.ListFiles("constraints")
		for _, file := range constraintFiles {
			path := "constraints/" + file
			content, _ := a.storage.ReadMarkdown(path)
			if len(content) > 0 {
				title := strings.TrimSuffix(file, ".md")
				id := extractIDFromFilename(file)
				events = append(events, &EventItem{
					ID:        id,
					Type:      "constraint",
					Title:     title,
					Content:   content,
					Timestamp: extractTimestampFromContent(content),
					Path:      path,
				})
			}
		}
	}

	// Sort by timestamp (newest first)
	sortEventsByTimestamp(events)

	return &EventListResponse{Events: events}, nil
}

// GraphResponse contains knowledge graph data
type GraphResponse struct {
	Nodes []*Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}

// Node represents a node in the knowledge graph
type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"` // "session", "decision", "discovery", "checkpoint", etc.
	Data  any    `json:"data,omitempty"`
}

// Edge represents an edge in the knowledge graph
type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"` // "contains", "relates_to"
}

// GetGraph returns the knowledge graph
func (a *API) GetGraph() (*GraphResponse, error) {
	var nodes []*Node
	var edges []*Edge

	// Get sessions
	sessions, _ := a.historyMgr.GetAllSessions("", "")
	for _, s := range sessions {
		nodes = append(nodes, &Node{
			ID:    s.ID,
			Label: s.Agent,
			Type:  "session",
			Data:  s,
		})
	}

	// Get decisions
	decisions, _ := a.storage.ListFiles("decisions")
	for _, file := range decisions {
		id := extractIDFromFilename(file)
		title := strings.TrimSuffix(file, ".md")
		nodes = append(nodes, &Node{
			ID:    id,
			Label: title,
			Type:  "decision",
		})
		// Connect to first session
		if len(sessions) > 0 {
			edges = append(edges, &Edge{
				Source: sessions[0].ID,
				Target: id,
				Type:   "contains",
			})
		}
	}

	// Get discoveries
	discoveries, _ := a.storage.ListFiles("discoveries")
	for _, file := range discoveries {
		id := extractIDFromFilename(file)
		title := strings.TrimSuffix(file, ".md")
		nodes = append(nodes, &Node{
			ID:    id,
			Label: title,
			Type:  "discovery",
		})
		if len(sessions) > 0 {
			edges = append(edges, &Edge{
				Source: sessions[0].ID,
				Target: id,
				Type:   "contains",
			})
		}
	}

	// Get checkpoints
	checkpoints, _ := a.storage.ListFiles("checkpoints")
	for _, file := range checkpoints {
		id := extractIDFromFilename(file)
		title := strings.TrimSuffix(file, ".md")
		nodes = append(nodes, &Node{
			ID:    id,
			Label: title,
			Type:  "checkpoint",
		})
		if len(sessions) > 0 {
			edges = append(edges, &Edge{
				Source: sessions[0].ID,
				Target: id,
				Type:   "contains",
			})
		}
	}

	return &GraphResponse{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

// SearchResponse contains search results
type SearchResponse struct {
	Results []*SearchResult `json:"results"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Excerpt string `json:"excerpt"`
	Path    string `json:"path"`
}

// Search searches across all entities
func (a *API) Search(query string) (*SearchResponse, error) {
	query = strings.ToLower(query)
	var results []*SearchResult

	// Search decisions
	decisions, _ := a.storage.ListFiles("decisions")
	for _, file := range decisions {
		path := "decisions/" + file
		content, _ := a.storage.ReadMarkdown(path)
		title := strings.TrimSuffix(file, ".md")

		if matches(title, content, query) {
			id := extractIDFromFilename(file)
			excerpt := extractExcerpt(content, 150)
			results = append(results, &SearchResult{
				ID:      id,
				Type:    "decision",
				Title:   title,
				Excerpt: excerpt,
				Path:    path,
			})
		}
	}

	// Search discoveries
	discoveries, _ := a.storage.ListFiles("discoveries")
	for _, file := range discoveries {
		path := "discoveries/" + file
		content, _ := a.storage.ReadMarkdown(path)
		title := strings.TrimSuffix(file, ".md")

		if matches(title, content, query) {
			id := extractIDFromFilename(file)
			excerpt := extractExcerpt(content, 150)
			results = append(results, &SearchResult{
				ID:      id,
				Type:    "discovery",
				Title:   title,
				Excerpt: excerpt,
				Path:    path,
			})
		}
	}

	// Search failures
	failures, _ := a.storage.ListFiles("failures")
	for _, file := range failures {
		path := "failures/" + file
		content, _ := a.storage.ReadMarkdown(path)
		title := strings.TrimSuffix(file, ".md")

		if matches(title, content, query) {
			id := extractIDFromFilename(file)
			excerpt := extractExcerpt(content, 150)
			results = append(results, &SearchResult{
				ID:      id,
				Type:    "failure",
				Title:   title,
				Excerpt: excerpt,
				Path:    path,
			})
		}
	}

	// Search constraints
	constraints, _ := a.storage.ListFiles("constraints")
	for _, file := range constraints {
		path := "constraints/" + file
		content, _ := a.storage.ReadMarkdown(path)
		title := strings.TrimSuffix(file, ".md")

		if matches(title, content, query) {
			id := extractIDFromFilename(file)
			excerpt := extractExcerpt(content, 150)
			results = append(results, &SearchResult{
				ID:      id,
				Type:    "constraint",
				Title:   title,
				Excerpt: excerpt,
				Path:    path,
			})
		}
	}

	// Search sessions
	sessions, _ := a.historyMgr.GetAllSessions("", "")
	for _, s := range sessions {
		if strings.Contains(strings.ToLower(s.Agent), query) {
			excerpt := fmt.Sprintf("Session %s (%s) on branch %s", s.ID[:8], s.Agent, s.Branch)
			results = append(results, &SearchResult{
				ID:      s.ID,
				Type:    "session",
				Title:   s.Agent,
				Excerpt: excerpt,
				Path:    fmt.Sprintf("sessions/%s", s.ID),
			})
		}
	}

	return &SearchResponse{Results: results}, nil
}

// Helper functions

func extractIDFromFilename(filename string) string {
	parts := strings.SplitN(filename, "-", 2)
	if len(parts) > 0 {
		return strings.TrimSuffix(parts[0], ".md")
	}
	return strings.TrimSuffix(filename, ".md")
}

func extractTimestampFromContent(content string) time.Time {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "timestamp") || strings.Contains(line, "Timestamp") {
			if idx := strings.Index(line, "202"); idx >= 0 && idx+19 <= len(line) {
				dateStr := line[idx : idx+19]
				if t, err := time.Parse("2006-01-02T15:04:05", dateStr); err == nil {
					return t
				}
			}
		}
	}
	return time.Now()
}

func extractExcerpt(content string, maxLen int) string {
	lines := strings.Split(content, "\n")
	excerpt := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "{") {
			excerpt = line
			break
		}
	}
	if len(excerpt) > maxLen {
		excerpt = excerpt[:maxLen] + "..."
	}
	return excerpt
}

func matches(title, content, query string) bool {
	contentLower := strings.ToLower(content)
	titleLower := strings.ToLower(title)
	return strings.Contains(titleLower, query) || strings.Contains(contentLower, query)
}

func sortEventsByTimestamp(events []*EventItem) {
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[j].Timestamp.After(events[i].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}
