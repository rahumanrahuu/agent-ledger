package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"agent-ledger/internal/memory"
)

// SearchRequest represents a memory search request
type SearchRequest struct {
	Query       string  `json:"query"`
	Type        string  `json:"type"`
	Threshold   float64 `json:"threshold"`
	Limit       int     `json:"limit"`
	TimeRange   string  `json:"time_range"`
}

// MemorySearchResponse represents search results
type MemorySearchResponse struct {
	Results []MemorySearchResultItem `json:"results"`
	Total   int                      `json:"total"`
	Query   string                   `json:"query"`
}

// MemorySearchResultItem represents a single search result
type MemorySearchResultItem struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	CreatedAt  string  `json:"created_at"`
	SessionID  string  `json:"session_id"`
	Excerpt    string  `json:"excerpt"`
}

// BriefingRequest represents a briefing request
type BriefingRequest struct {
	Task       string `json:"task"`
	SessionID  string `json:"session_id"`
}

// BriefingResponse represents a generated briefing
type BriefingResponse struct {
	Task                 string   `json:"task"`
	TechStack            []string `json:"tech_stack"`
	Architecture         string   `json:"architecture"`
	Constraints          []string `json:"constraints"`
	Decisions            []string `json:"decisions"`
	Risks                []string `json:"risks"`
	Failures             []string `json:"failures"`
	NextSteps            []string `json:"next_steps"`
	EstimatedDuration    string   `json:"estimated_duration"`
	EstimatedCost        int      `json:"estimated_cost"`
}

// HandleMemorySearch handles memory search requests
func (a *API) HandleMemorySearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if memory manager is available
	if a.memoryMgr == nil {
		http.Error(w, "Memory system not available", http.StatusServiceUnavailable)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	memType := r.URL.Query().Get("type")
	if memType == "" {
		memType = "all"
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	thresholdStr := r.URL.Query().Get("threshold")
	threshold := 0.6
	if thresholdStr != "" {
		if t, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
			threshold = t
		}
	}

	// Perform search
	results, err := a.memoryMgr.Search(query, memType, limit)
	if err != nil {
		http.Error(w, "Search failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter by threshold and convert results
	var filteredResults []MemorySearchResultItem
	for _, result := range results {
		if result.Score >= threshold {
			excerpt := result.Memory.Content
			if len(excerpt) > 120 {
				excerpt = excerpt[:120] + "..."
			}

			filteredResults = append(filteredResults, MemorySearchResultItem{
				ID:        result.Memory.ID,
				Type:      result.Memory.Type,
				Title:     result.Memory.Title,
				Content:   result.Memory.Content,
				Score:     result.Score,
				CreatedAt: result.Memory.CreatedAt.Format("2006-01-02 15:04:05"),
				SessionID: result.Memory.SessionID,
				Excerpt:   excerpt,
			})
		}
	}

	response := MemorySearchResponse{
		Results: filteredResults,
		Total:   len(filteredResults),
		Query:   query,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleBriefing generates a context briefing for a task
func (a *API) HandleBriefing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if memory manager is available
	if a.memoryMgr == nil {
		http.Error(w, "Memory system not available", http.StatusServiceUnavailable)
		return
	}

	task := r.URL.Query().Get("task")
	if task == "" {
		http.Error(w, "Task parameter is required", http.StatusBadRequest)
		return
	}

	_ = r.URL.Query().Get("session_id") // sessionID declared but not used - remove this line

	// Get relevant memories for this task
	decisions, _ := a.memoryMgr.Search(task+" decision", "decision", 5)
	discoveries, _ := a.memoryMgr.Search(task+" discovery", "discovery", 5)
	constraints, _ := a.memoryMgr.Search(task, "constraint", 10)
	failures, _ := a.memoryMgr.Search(task+" failure", "failure", 3)

	// Extract text from memories
	var techStack, architecture, decisionList, riskList, nextSteps, failureList []string

	for _, d := range discoveries {
		if len(techStack) < 5 {
			techStack = append(techStack, d.Memory.Title)
		}
	}

	for _, d := range decisions {
		if len(decisionList) < 5 {
			decisionList = append(decisionList, d.Memory.Title)
		}
	}

	var constraintList []string
	for _, c := range constraints {
		if len(constraintList) < 5 {
			constraintList = append(constraintList, c.Memory.Title)
		}
	}

	for _, f := range failures {
		if len(failureList) < 3 {
			failureList = append(failureList, f.Memory.Title)
		}
	}

	// Generate estimated duration based on task complexity
	estimatedDuration := "45-60 minutes"
	if strings.Contains(strings.ToLower(task), "small") || strings.Contains(strings.ToLower(task), "quick") {
		estimatedDuration = "15-30 minutes"
	} else if strings.Contains(strings.ToLower(task), "large") || strings.Contains(strings.ToLower(task), "complex") {
		estimatedDuration = "2-4 hours"
	}

	briefing := BriefingResponse{
		Task:              task,
		TechStack:         techStack,
		Architecture:      "See decisions and discoveries for details",
		Constraints:       constraintList,
		Decisions:         decisionList,
		Risks:             riskList,
		Failures:          failureList,
		NextSteps:         nextSteps,
		EstimatedDuration: estimatedDuration,
		EstimatedCost:     2500, // Default estimate
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(briefing)
}

// HandleMemoryList lists all memories
func (a *API) HandleMemoryList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	memType := r.URL.Query().Get("type")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	memories, err := a.memoryMgr.List(memType, limit)
	if err != nil {
		http.Error(w, "List failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"memories": memories,
		"total":    len(memories),
	})
}

// HandleMemoryAdd adds a new memory
func (a *API) HandleMemoryAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var mem memory.Memory
	if err := json.NewDecoder(r.Body).Decode(&mem); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.memoryMgr.Add(mem); err != nil {
		http.Error(w, "Add failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": mem.ID, "status": "created"})
}
