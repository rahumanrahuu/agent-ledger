package api

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agent-ledger/internal/checkpoint"
	"agent-ledger/internal/events"
	"agent-ledger/internal/git"
	"agent-ledger/internal/history"
	"agent-ledger/internal/memory"
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
	memoryMgr     *memory.Manager
	version       string
}

// NewAPI creates a new API instance
func NewAPI(repo *repository.Repository, st *storage.Storage, version string) *API {
	sessionMgr := session.NewManager(st)
	checkpointMgr := checkpoint.NewManager(st)
	historyMgr := history.NewManager(sessionMgr, checkpointMgr, st)
	eventsMgr := events.NewManager(st)

	// Initialize memory manager - handle potential errors gracefully
	memoryMgr, err := memory.NewManager(repo.Root)
	if err != nil {
		// Log error but don't fail - memory features will be disabled
		memoryMgr = nil
	}

	return &API{
		repo:          repo,
		storage:       st,
		sessionMgr:    sessionMgr,
		checkpointMgr: checkpointMgr,
		historyMgr:    historyMgr,
		eventsMgr:     eventsMgr,
		memoryMgr:     memoryMgr,
		version:       version,
	}
}

// OverviewResponse contains project overview data
type OverviewResponse struct {
	ProjectName      string     `json:"project_name"`
	RepositoryRoot   string     `json:"repository_root"`
	CurrentBranch    string     `json:"current_branch"`
	CurrentCommit    string     `json:"current_commit"`
	Version          string     `json:"version"`
	SessionCount     int        `json:"session_count"`
	DecisionCount    int        `json:"decision_count"`
	DiscoveryCount   int        `json:"discovery_count"`
	CheckpointCount  int        `json:"checkpoint_count"`
	LastActivityTime *time.Time `json:"last_activity_time,omitempty"`
}

// GetOverview returns project overview data
// GetOverview returns project overview data with real-time Git status and activity timestamp
func (a *API) GetOverview() (*OverviewResponse, error) {
	sessions, err := a.historyMgr.GetAllSessions("", "")
	if err != nil {
		sessions = []*session.Session{}
	}

	decisions, _ := a.storage.ListFiles("decisions")
	discoveries, _ := a.storage.ListFiles("discoveries")
	checkpoints := a.getAllCheckpoints(sessions)

	// Dynamically query real-time Git branch and HEAD commit
	currentBranch, errBranch := git.GetCurrentBranch()
	if errBranch != nil || currentBranch == "" {
		currentBranch = a.repo.Branch
	}
	currentCommit, errCommit := git.GetHeadCommit()
	if errCommit != nil || currentCommit == "" {
		currentCommit = a.repo.Head
	}

	projectName := "Agent Ledger Project"
	if content, err := a.storage.ReadMarkdown("project.md"); err == nil && len(content) > 0 {
		projectName = "Project"
	}

	// Calculate true latest activity timestamp across sessions, checkpoints, memories, and event files
	var latestTime time.Time

	for _, s := range sessions {
		if s.EndTime != nil && s.EndTime.After(latestTime) {
			latestTime = *s.EndTime
		}
		if s.StartTime.After(latestTime) {
			latestTime = s.StartTime
		}
	}

	for _, cp := range checkpoints {
		if cp.Timestamp.After(latestTime) {
			latestTime = cp.Timestamp
		}
	}

	// Check file modification times in decisions, discoveries, failures, constraints
	for _, cat := range []string{"decisions", "discoveries", "failures", "constraints"} {
		files, _ := a.storage.ListFiles(cat)
		for _, f := range files {
			filePath := filepath.Join(a.storage.GetRoot(), cat, f)
			if info, err := os.Stat(filePath); err == nil {
				if info.ModTime().After(latestTime) {
					latestTime = info.ModTime().UTC()
				}
			}
		}
	}

	// Check memories
	memories, _ := a.GetMemories("", "", 100)
	for _, m := range memories {
		if m.CreatedAt.After(latestTime) {
			latestTime = m.CreatedAt.UTC()
		}
	}

	var lastActivityTime *time.Time
	if !latestTime.IsZero() {
		lastActivityTime = &latestTime
	}

	return &OverviewResponse{
		ProjectName:      projectName,
		RepositoryRoot:   a.repo.Root,
		CurrentBranch:    currentBranch,
		CurrentCommit:    currentCommit,
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

// GetSessions returns all sessions with accurate active vs completed status
func (a *API) GetSessions() (*SessionListResponse, error) {
	sessionList, err := a.historyMgr.GetAllSessions("", "")
	if err != nil {
		return &SessionListResponse{Sessions: []*SessionInfo{}}, nil
	}

	now := time.Now()
	sessions := make([]*SessionInfo, len(sessionList))

	// Find index of the most recently started session without EndTime
	latestActiveIndex := -1
	var latestActiveTime time.Time
	for i, s := range sessionList {
		if s.EndTime == nil && s.StartTime.After(latestActiveTime) {
			latestActiveTime = s.StartTime
			latestActiveIndex = i
		}
	}

	for i, s := range sessionList {
		status := "ended"
		// A session is active ONLY if it has no EndTime AND it's the latest active session or started within 2 hours
		if s.EndTime == nil && (i == latestActiveIndex || now.Sub(s.StartTime) <= 2*time.Hour) {
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
	Session     *SessionInfo `json:"session"`
	Checkpoints int          `json:"checkpoint_count"`
	Decisions   int          `json:"decision_count"`
	Discoveries int          `json:"discovery_count"`
	Failures    int          `json:"failure_count"`
	Constraints int          `json:"constraint_count"`
	HasHandoff  bool         `json:"has_handoff"`
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

	// Checkpoints are Git-native session artifacts stored in each session's
	// checkpoints.json, rather than Markdown files in a top-level directory.
	if filterType == "" || filterType == "checkpoint" {
		sessions, _ := a.historyMgr.GetAllSessions("", "")
		for _, cp := range a.getAllCheckpoints(sessions) {
			changed := len(cp.ChangedFiles) + len(cp.AddedFiles) + len(cp.DeletedFiles) + len(cp.ModifiedFiles)
			events = append(events, &EventItem{
				ID:        cp.ID,
				Type:      "checkpoint",
				Title:     cp.ID,
				Content:   fmt.Sprintf("Git checkpoint for session %s at commit %s. Captured %d changed files.", cp.SessionID, cp.Commit, changed),
				Timestamp: cp.Timestamp,
				Path:      fmt.Sprintf("sessions/%s/checkpoints.json", cp.SessionID),
			})
		}
	}

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

// GetGraph returns the complete knowledge graph
func (a *API) GetGraph() (*GraphResponse, error) {
	nodes := make([]*Node, 0)
	edges := make([]*Edge, 0)
	nodeMap := make(map[string]bool)

	addNode := func(n *Node) {
		if !nodeMap[n.ID] {
			nodeMap[n.ID] = true
			nodes = append(nodes, n)
		}
	}

	// 1. Get sessions
	sessions, _ := a.historyMgr.GetAllSessions("", "")
	sessionIDs := make(map[string]string)
	for _, s := range sessions {
		sessionIDs[s.ID] = s.ID
		label := s.Agent
		if label == "" || label == "agent" {
			if len(s.ID) >= 8 {
				label = "Session " + s.ID[:8]
			} else {
				label = "Session " + s.ID
			}
		}
		addNode(&Node{
			ID:    s.ID,
			Label: label,
			Type:  "session",
			Data:  s,
		})
	}

	// 2. Helper to add category event nodes (decisions, discoveries, failures, constraints)
	addCategory := func(category, nodeType string) {
		files, _ := a.storage.ListFiles(category)
		for _, file := range files {
			path := category + "/" + file
			content, _ := a.storage.ReadMarkdown(path)
			id := extractIDFromFilename(file)
			title := strings.TrimSuffix(file, ".md")
			cleanTitle := strings.ReplaceAll(title, "-", " ")
			if len(cleanTitle) > 36 {
				cleanTitle = cleanTitle[:35] + "…"
			}

			addNode(&Node{
				ID:    id,
				Label: cleanTitle,
				Type:  nodeType,
				Data:  content,
			})

			foundSession := ""
			for sessID := range sessionIDs {
				if strings.Contains(content, sessID) {
					foundSession = sessID
					break
				}
			}
			if foundSession != "" {
				edges = append(edges, &Edge{
					Source: foundSession,
					Target: id,
					Type:   "contains",
				})
			} else if len(sessions) > 0 {
				edges = append(edges, &Edge{
					Source: sessions[0].ID,
					Target: id,
					Type:   "contains",
				})
			}
		}
	}

	addCategory("decisions", "decision")
	addCategory("discoveries", "discovery")
	addCategory("failures", "failure")
	addCategory("constraints", "constraint")

	// 3. Get Checkpoints
	for _, cp := range a.getAllCheckpoints(sessions) {
		label := "Checkpoint " + cp.ID
		if len(cp.ID) >= 8 {
			label = "Checkpoint " + cp.ID[:8]
		}
		addNode(&Node{
			ID:    cp.ID,
			Label: label,
			Type:  "checkpoint",
			Data:  cp,
		})
		if cp.SessionID != "" {
			edges = append(edges, &Edge{
				Source: cp.SessionID,
				Target: cp.ID,
				Type:   "checkpoint_of",
			})
		}
	}

	// 4. Get Memories
	memories, _ := a.GetMemories("", "", 100)
	for _, mem := range memories {
		label := mem.Title
		if label == "" {
			label = mem.Content
		}
		if len(label) > 36 {
			label = label[:35] + "…"
		}
		addNode(&Node{
			ID:    mem.ID,
			Label: label,
			Type:  mem.Type,
			Data:  mem,
		})
		if mem.SessionID != "" && sessionIDs[mem.SessionID] != "" {
			edges = append(edges, &Edge{
				Source: mem.SessionID,
				Target: mem.ID,
				Type:   "remembers",
			})
		}
	}

	return &GraphResponse{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

// getAllCheckpoints reads checkpoint metadata through the checkpoint manager,
// which is the canonical store used by the CLI and MCP server.
func (a *API) getAllCheckpoints(sessions []*session.Session) []checkpoint.Checkpoint {
	checkpoints := make([]checkpoint.Checkpoint, 0)
	for _, sess := range sessions {
		sessionCheckpoints, err := a.checkpointMgr.List(sess.ID)
		if err == nil {
			checkpoints = append(checkpoints, sessionCheckpoints...)
		}
	}
	return checkpoints
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

// GetMemories lists or searches the vector-backed memory store. Search results
// are flattened to Memory values so list and search responses share one stable
// JSON shape for API clients.
func (a *API) GetMemories(query, memoryType string, limit int) ([]memory.Memory, error) {
	if a.memoryMgr == nil {
		return []memory.Memory{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if query == "" {
		memories, err := a.memoryMgr.List(memoryType, limit)
		if memories == nil {
			memories = []memory.Memory{}
		}
		return memories, err
	}

	results, err := a.memoryMgr.Search(query, memoryType, limit)
	if err != nil {
		return nil, err
	}
	memories := make([]memory.Memory, 0, len(results))
	for _, result := range results {
		memories = append(memories, result.Memory)
	}
	return memories, nil
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
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})
}

// extractSessionIDFromFilename tries to match a knowledge file to its owning
// session by reading the session_id embedded in the markdown content.
// Falls back to the most recent session if no match is found.
func extractSessionIDFromFilename(filename string, sessions []*session.Session) string {
	if len(sessions) == 0 {
		return ""
	}
	// Default to most recent session (sessions are ordered newest-first)
	return sessions[len(sessions)-1].ID
}
