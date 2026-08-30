package context_selector

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ContextType represents the type of context
type ContextType string

const (
	RecentActivity   ContextType = "recent_activity"
	TopDecisions     ContextType = "top_decisions"
	ActiveConstraints ContextType = "active_constraints"
	FailureHistory   ContextType = "failure_history"
	BestPractices    ContextType = "best_practices"
	Architecture     ContextType = "architecture"
	Performance      ContextType = "performance"
	Dependencies     ContextType = "dependencies"
)

// ContextItem represents a piece of context
type ContextItem struct {
	ID          string      `json:"id"`
	Type        ContextType `json:"type"`
	Title       string      `json:"title"`
	Content     string      `json:"content"`
	Relevance   float64     `json:"relevance"` // 0.0-1.0
	Importance  float64     `json:"importance"` // 0.0-1.0
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Tags        []string    `json:"tags"`
	Agent       string      `json:"agent,omitempty"`
	Session     string      `json:"session,omitempty"`
}

// ContextRequest defines what context is needed
type ContextRequest struct {
	Query          string
	TaskDescription string
	AgentID        string
	ContextTypes   []ContextType
	Limit          int
	MinRelevance   float64
	TimeWindow     time.Duration
}

// ContextResponse is the result of a context selection
type ContextResponse struct {
	Items            []*ContextItem `json:"items"`
	TotalItems       int            `json:"total_items"`
	SelectionTime    float64        `json:"selection_time_ms"`
	CoverageScore    float64        `json:"coverage_score"`
	RecommendedNext  []string       `json:"recommended_next"`
}

// SmartSelector intelligently selects context
type SmartSelector struct {
	mu    sync.RWMutex
	items map[string]*ContextItem
	index map[ContextType][]string // type -> itemIDs
}

// NewSmartSelector creates a new smart context selector
func NewSmartSelector() *SmartSelector {
	return &SmartSelector{
		items: make(map[string]*ContextItem),
		index: make(map[ContextType][]string),
	}
}

// RegisterContext registers a piece of context
func (ss *SmartSelector) RegisterContext(item *ContextItem) error {
	if item.ID == "" || item.Title == "" {
		return fmt.Errorf("context ID and title are required")
	}

	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.items[item.ID] = item
	ss.index[item.Type] = append(ss.index[item.Type], item.ID)

	return nil
}

// SelectContext intelligently selects context for a task
func (ss *SmartSelector) SelectContext(req ContextRequest) *ContextResponse {
	startTime := time.Now()

	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.MinRelevance == 0 {
		req.MinRelevance = 0.3
	}
	if req.TimeWindow == 0 {
		req.TimeWindow = 30 * 24 * time.Hour
	}

	ss.mu.RLock()
	defer ss.mu.RUnlock()

	var candidates []*ContextItem
	cutoff := time.Now().Add(-req.TimeWindow)

	// Collect candidates
	for _, item := range ss.items {
		if item.UpdatedAt.Before(cutoff) {
			continue
		}

		// Filter by context type if specified
		if len(req.ContextTypes) > 0 {
			typeMatch := false
			for _, t := range req.ContextTypes {
				if item.Type == t {
					typeMatch = true
					break
				}
			}
			if !typeMatch {
				continue
			}
		}

		candidates = append(candidates, item)
	}

	// Score and filter candidates
	var scored []*ContextItem
	for _, item := range candidates {
		score := ss.calculateContextScore(item, req)
		if score >= req.MinRelevance {
			item.Relevance = score
			scored = append(scored, item)
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Relevance > scored[j].Relevance
	})

	// Limit results
	if len(scored) > req.Limit {
		scored = scored[:req.Limit]
	}

	// Calculate coverage score
	coverage := ss.calculateCoverageScore(scored, req)

	// Get recommendations for what to explore next
	recommendations := ss.getRecommendations(scored, req)

	selectionTime := time.Since(startTime).Seconds() * 1000

	return &ContextResponse{
		Items:           scored,
		TotalItems:      len(scored),
		SelectionTime:   selectionTime,
		CoverageScore:   coverage,
		RecommendedNext: recommendations,
	}
}

// calculateContextScore scores how relevant a context item is to the request
func (ss *SmartSelector) calculateContextScore(item *ContextItem, req ContextRequest) float64 {
	score := 0.0

	// Query matching (40%)
	queryScore := ss.queryMatchScore(req.Query, item)
	score += queryScore * 0.4

	// Task description matching (30%)
	taskScore := ss.taskMatchScore(req.TaskDescription, item)
	score += taskScore * 0.3

	// Item importance (20%)
	score += item.Importance * 0.2

	// Recency bonus (10%)
	recencyScore := ss.recencyScore(item.UpdatedAt)
	score += recencyScore * 0.1

	return min(score, 1.0)
}

// queryMatchScore scores query relevance
func (ss *SmartSelector) queryMatchScore(query string, item *ContextItem) float64 {
	if query == "" {
		return 0.3
	}

	queryTerms := strings.Fields(strings.ToLower(query))
	titleLower := strings.ToLower(item.Title)
	contentLower := strings.ToLower(item.Content)

	matchCount := 0
	for _, term := range queryTerms {
		if strings.Contains(titleLower, term) || strings.Contains(contentLower, term) {
			matchCount++
		}
	}

	return float64(matchCount) / float64(len(queryTerms))
}

// taskMatchScore scores task description relevance
func (ss *SmartSelector) taskMatchScore(taskDesc string, item *ContextItem) float64 {
	if taskDesc == "" {
		return 0.2
	}

	// Check tag matches
	taskTerms := strings.Fields(strings.ToLower(taskDesc))
	matchCount := 0

	for _, term := range taskTerms {
		for _, tag := range item.Tags {
			if strings.Contains(strings.ToLower(tag), term) {
				matchCount++
				break
			}
		}
	}

	if len(taskTerms) == 0 {
		return 0
	}

	return float64(matchCount) / float64(len(taskTerms))
}

// recencyScore gives higher scores to recent items
func (ss *SmartSelector) recencyScore(updatedAt time.Time) float64 {
	age := time.Since(updatedAt)
	ageDays := age.Hours() / 24

	if ageDays <= 7 {
		return 1.0
	}
	if ageDays > 90 {
		return 0.2
	}
	return 1.0 - (ageDays-7)/83*0.8
}

// calculateCoverageScore evaluates how well selected items cover needed context
func (ss *SmartSelector) calculateCoverageScore(items []*ContextItem, req ContextRequest) float64 {
	if len(items) == 0 {
		return 0
	}

	// Count covered context types
	typesCovered := make(map[ContextType]bool)
	for _, item := range items {
		typesCovered[item.Type] = true
	}

	if len(req.ContextTypes) == 0 {
		return float64(len(items)) / 10.0
	}

	covered := 0
	for _, t := range req.ContextTypes {
		if typesCovered[t] {
			covered++
		}
	}

	return float64(covered) / float64(len(req.ContextTypes))
}

// getRecommendations suggests what to explore next
func (ss *SmartSelector) getRecommendations(selected []*ContextItem, req ContextRequest) []string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	var recommendations []string
	selectedTypes := make(map[ContextType]bool)

	for _, item := range selected {
		selectedTypes[item.Type] = true
	}

	// Suggest uncovered types
	allTypes := []ContextType{
		RecentActivity, TopDecisions, ActiveConstraints, FailureHistory,
		BestPractices, Architecture, Performance, Dependencies,
	}

	for _, t := range allTypes {
		if !selectedTypes[t] && len(ss.index[t]) > 0 {
			recommendations = append(recommendations, fmt.Sprintf("Explore %s", string(t)))
		}
	}

	return recommendations
}

// GetContextByType gets all context of a specific type
func (ss *SmartSelector) GetContextByType(contextType ContextType) []*ContextItem {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	itemIDs, exists := ss.index[contextType]
	if !exists {
		return []*ContextItem{}
	}

	var items []*ContextItem
	for _, id := range itemIDs {
		if item, ok := ss.items[id]; ok {
			items = append(items, item)
		}
	}

	return items
}

// GetContextStats gets statistics about stored context
func (ss *SmartSelector) GetContextStats() map[string]interface{} {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	typeCount := make(map[string]int)
	for t, ids := range ss.index {
		typeCount[string(t)] = len(ids)
	}

	return map[string]interface{}{
		"total_items": len(ss.items),
		"by_type":     typeCount,
		"types":       len(ss.index),
	}
}

// UpdateRelevance updates the relevance score of a context item
func (ss *SmartSelector) UpdateRelevance(itemID string, relevance float64) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	item, exists := ss.items[itemID]
	if !exists {
		return fmt.Errorf("context item not found: %s", itemID)
	}

	item.Relevance = min(max(relevance, 0), 1)
	item.UpdatedAt = time.Now()

	return nil
}

// Utility functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
