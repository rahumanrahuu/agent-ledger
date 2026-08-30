package api

import (
	"fmt"
	"sort"
	"time"
)

// AnalyticsMetrics contains comprehensive project metrics
type AnalyticsMetrics struct {
	ProjectStats  ProjectStats    `json:"project_stats"`
	ActivityTrend []ActivityPoint `json:"activity_trend"`
	AgentStats    []AgentMetrics   `json:"agent_stats"`
	ModelStats    []ModelMetrics   `json:"model_stats"`
	EventTypeStats EventTypeCounts `json:"event_type_stats"`
	DecisionMetrics DecisionMetrics `json:"decision_metrics"`
}

// ProjectStats contains overall project statistics
type ProjectStats struct {
	TotalSessions       int     `json:"total_sessions"`
	ActiveSessions      int     `json:"active_sessions"`
	TotalRecords        int     `json:"total_records"`
	AverageSessionDays  float64 `json:"average_session_duration_days"`
	ProjectAgeInDays    int     `json:"project_age_in_days"`
	LastActivityTime    string  `json:"last_activity_time"`
}

// ActivityPoint represents activity at a point in time
type ActivityPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// AgentMetrics contains per-agent statistics
type AgentMetrics struct {
	Name            string `json:"name"`
	SessionCount    int    `json:"session_count"`
	RecordCount     int    `json:"record_count"`
	LastActive      string `json:"last_active"`
	AverageRecords  int    `json:"average_records_per_session"`
}

// ModelMetrics contains per-model statistics
type ModelMetrics struct {
	Name         string `json:"name"`
	UsageCount   int    `json:"usage_count"`
	SessionCount int    `json:"session_count"`
}

// EventTypeCounts contains counts by event type
type EventTypeCounts struct {
	Decisions    int `json:"decisions"`
	Discoveries  int `json:"discoveries"`
	Failures     int `json:"failures"`
	Constraints  int `json:"constraints"`
	Checkpoints  int `json:"checkpoints"`
}

// DecisionMetrics contains decision-specific metrics
type DecisionMetrics struct {
	Total              int    `json:"total"`
	ArchitecturalCount int    `json:"architectural_count"`
	ImplementationCount int   `json:"implementation_count"`
	CommonPatterns     []string `json:"common_patterns"`
}

// GetAnalytics returns comprehensive project analytics
func (a *API) GetAnalytics() (*AnalyticsMetrics, error) {
	sessions, _ := a.historyMgr.GetAllSessions("", "")

	decisions, _ := a.storage.ListFiles("decisions")
	discoveries, _ := a.storage.ListFiles("discoveries")
	failures, _ := a.storage.ListFiles("failures")
	constraints, _ := a.storage.ListFiles("constraints")

	checkpoints := a.getAllCheckpoints(sessions)

	activeSessions := 0
	var oldestTime time.Time
	var latestTime time.Time

	agentMap := make(map[string]*AgentMetrics)
	modelMap := make(map[string]*ModelMetrics)

	for i, s := range sessions {
		if s.EndTime == nil {
			activeSessions++
		}

		if i == 0 || s.StartTime.Before(oldestTime) {
			oldestTime = s.StartTime
		}
		if i == 0 || s.StartTime.After(latestTime) {
			latestTime = s.StartTime
			if s.EndTime != nil && s.EndTime.After(latestTime) {
				latestTime = *s.EndTime
			}
		}

		agent := s.Agent
		if agent == "" {
			agent = "unspecified"
		}
		if _, ok := agentMap[agent]; !ok {
			agentMap[agent] = &AgentMetrics{Name: agent}
		}
		agentMap[agent].SessionCount++

		model := s.Model
		if model == "" {
			model = "unspecified"
		}
		if _, ok := modelMap[model]; !ok {
			modelMap[model] = &ModelMetrics{Name: model}
		}
		modelMap[model].UsageCount++
		modelMap[model].SessionCount++
	}

	agentStats := make([]AgentMetrics, 0)
	for _, stat := range agentMap {
		if stat.SessionCount > 0 {
			stat.AverageRecords = (len(decisions) + len(discoveries) + len(failures) + len(constraints)) / stat.SessionCount
		}
		agentStats = append(agentStats, *stat)
	}
	sort.Slice(agentStats, func(i, j int) bool {
		return agentStats[i].SessionCount > agentStats[j].SessionCount
	})

	modelStats := make([]ModelMetrics, 0)
	for _, stat := range modelMap {
		modelStats = append(modelStats, *stat)
	}
	sort.Slice(modelStats, func(i, j int) bool {
		return modelStats[i].UsageCount > modelStats[j].UsageCount
	})

	projectAgeDays := 0
	if !oldestTime.IsZero() {
		projectAgeDays = int(time.Since(oldestTime).Hours() / 24)
	}

	avgSessionDays := 0.0
	if len(sessions) > 0 && !oldestTime.IsZero() && !latestTime.IsZero() {
		avgSessionDays = time.Since(oldestTime).Hours() / 24 / float64(len(sessions))
	}

	lastActivity := ""
	if !latestTime.IsZero() {
		lastActivity = latestTime.Format(time.RFC3339)
	}

	eventTypes := EventTypeCounts{
		Decisions:   len(decisions),
		Discoveries: len(discoveries),
		Failures:    len(failures),
		Constraints: len(constraints),
		Checkpoints: len(checkpoints),
	}

	return &AnalyticsMetrics{
		ProjectStats: ProjectStats{
			TotalSessions:     len(sessions),
			ActiveSessions:    activeSessions,
			TotalRecords:      eventTypes.Decisions + eventTypes.Discoveries + eventTypes.Failures + eventTypes.Constraints,
			AverageSessionDays: avgSessionDays,
			ProjectAgeInDays:  projectAgeDays,
			LastActivityTime: lastActivity,
		},
		AgentStats:     agentStats,
		ModelStats:     modelStats,
		EventTypeStats: eventTypes,
		DecisionMetrics: DecisionMetrics{
			Total: len(decisions),
		},
	}, nil
}

// AdvancedSearchQuery represents advanced search parameters
type AdvancedSearchQuery struct {
	Keywords     []string `json:"keywords"`
	RecordTypes  []string `json:"record_types"`
	SessionID    string   `json:"session_id,omitempty"`
	Agent        string   `json:"agent,omitempty"`
	Model        string   `json:"model,omitempty"`
	DateRangeStart string  `json:"date_range_start,omitempty"`
	DateRangeEnd  string   `json:"date_range_end,omitempty"`
	Limit        int      `json:"limit"`
}

// AdvancedSearchResults represents results from an advanced search
type AdvancedSearchResults struct {
	Query       *AdvancedSearchQuery `json:"query"`
	Results     []*SearchResult      `json:"results"`
	Total       int                  `json:"total"`
	ExecutionMs int64               `json:"execution_ms"`
}

// PerformAdvancedSearch executes a complex query across the ledger
func (a *API) PerformAdvancedSearch(q *AdvancedSearchQuery) (*AdvancedSearchResults, error) {
	start := time.Now()

	if q.Limit == 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	results := make([]*SearchResult, 0)

	// If we have a memory manager, use it for keyword search
	if a.memoryMgr != nil && len(q.Keywords) > 0 {
		for _, keyword := range q.Keywords {
			memType := ""
			if len(q.RecordTypes) == 1 {
				memType = q.RecordTypes[0]
			}

			searchResults, err := a.memoryMgr.Search(keyword, memType, q.Limit)
			if err == nil {
				for _, sr := range searchResults {
					results = append(results, &SearchResult{
						ID:      sr.Memory.ID,
						Type:    sr.Memory.Type,
						Title:   sr.Memory.Title,
						Excerpt: sr.Memory.Content,
						Path:    sr.Memory.Path,
					})
				}
			}
		}
	}

	// Deduplicate results
	seen := make(map[string]bool)
	dedupResults := make([]*SearchResult, 0)
	for _, r := range results {
		if !seen[r.ID] {
			seen[r.ID] = true
			dedupResults = append(dedupResults, r)
		}
	}

	if len(dedupResults) > q.Limit {
		dedupResults = dedupResults[:q.Limit]
	}

	executionMs := time.Since(start).Milliseconds()

	return &AdvancedSearchResults{
		Query:       q,
		Results:     dedupResults,
		Total:       len(dedupResults),
		ExecutionMs: executionMs,
	}, nil
}

// RelatedItemsQuery represents parameters for finding related items
type RelatedItemsQuery struct {
	ItemID         string   `json:"item_id"`
	RelationType   string   `json:"relation_type"` // "similar", "same_session", "same_agent"
	Limit          int      `json:"limit"`
}

// RelatedItems contains related items by category
type RelatedItems struct {
	QueryItem *SearchResult   `json:"query_item"`
	Similar   []*SearchResult `json:"similar"`
	FromSession []*SearchResult `json:"from_session"`
	FromAgent []*SearchResult `json:"from_agent"`
}

// GetRelatedItems finds items related to a given item
func (a *API) GetRelatedItems(q *RelatedItemsQuery) (*RelatedItems, error) {
	if q.Limit == 0 {
		q.Limit = 10
	}

	if q.Limit > 50 {
		q.Limit = 50
	}

	// Get the query item
	var queryItem *SearchResult

	if a.memoryMgr != nil {
		mem, err := a.memoryMgr.Get(q.ItemID)
		if err == nil {
			queryItem = &SearchResult{
				ID:      mem.ID,
				Type:    mem.Type,
				Title:   mem.Title,
				Excerpt: mem.Content,
				Path:    mem.Path,
			}
		}
	}

	result := &RelatedItems{
		QueryItem: queryItem,
		Similar:   make([]*SearchResult, 0),
		FromSession: make([]*SearchResult, 0),
		FromAgent: make([]*SearchResult, 0),
	}

	if queryItem == nil {
		return result, fmt.Errorf("item not found")
	}

	// Find similar items using memory manager
	if a.memoryMgr != nil {
		similar, err := a.memoryMgr.Search(queryItem.Title, "", q.Limit+1)
		if err == nil {
			for _, sr := range similar {
				if sr.Memory.ID != q.ItemID {
					result.Similar = append(result.Similar, &SearchResult{
						ID:      sr.Memory.ID,
						Type:    sr.Memory.Type,
						Title:   sr.Memory.Title,
						Excerpt: sr.Memory.Content,
						Path:    sr.Memory.Path,
					})
				}
				if len(result.Similar) >= q.Limit {
					break
				}
			}
		}
	}

	return result, nil
}
