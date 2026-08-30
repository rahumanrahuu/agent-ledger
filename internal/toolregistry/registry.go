package toolregistry

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ToolDefinition defines a tool that can be invoked
type ToolDefinition struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Tags         []string               `json:"tags"`
	InputSchema  map[string]interface{} `json:"input_schema"`
	OutputSchema map[string]interface{} `json:"output_schema"`
	Owner        string                 `json:"owner"` // agent or service ID
	Version      string                 `json:"version"`
	Status       string                 `json:"status"` // active, deprecated, beta
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	UsageCount   int64                  `json:"usage_count"`
	SuccessRate  float64                `json:"success_rate"` // 0.0-1.0
	AvgLatencyMs float64                `json:"avg_latency_ms"`
	Examples     []ToolExample          `json:"examples"`
}

// ToolExample shows how to use a tool
type ToolExample struct {
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output"`
}

// ToolUsage tracks tool usage
type ToolUsage struct {
	ID        string    `json:"id"`
	ToolID    string    `json:"tool_id"`
	AgentID   string    `json:"agent_id"`
	Success   bool      `json:"success"`
	LatencyMs float64   `json:"latency_ms"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// Registry manages available tools
type Registry struct {
	mu      sync.RWMutex
	tools   map[string]*ToolDefinition
	index   map[string][]string // category -> toolIDs, tag -> toolIDs
	usage   map[string]*ToolUsage
	tagIdx  map[string][]string
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools:   make(map[string]*ToolDefinition),
		index:   make(map[string][]string),
		usage:   make(map[string]*ToolUsage),
		tagIdx:  make(map[string][]string),
	}
}

// Register registers a new tool
func (r *Registry) Register(tool *ToolDefinition) error {
	if tool.Name == "" || tool.Description == "" {
		return fmt.Errorf("tool name and description are required")
	}

	tool.ID = uuid.New().String()
	tool.CreatedAt = time.Now()
	tool.UpdatedAt = time.Now()
	if tool.Status == "" {
		tool.Status = "active"
	}
	if tool.Version == "" {
		tool.Version = "1.0.0"
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools[tool.ID] = tool

	// Index by category
	if tool.Category != "" {
		r.index[tool.Category] = append(r.index[tool.Category], tool.ID)
	}

	// Index by tags
	for _, tag := range tool.Tags {
		r.tagIdx[tag] = append(r.tagIdx[tag], tool.ID)
	}

	return nil
}

// Unregister removes a tool
func (r *Registry) Unregister(toolID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tool, exists := r.tools[toolID]
	if !exists {
		return fmt.Errorf("tool not found: %s", toolID)
	}

	delete(r.tools, toolID)

	// Remove from category index
	if tool.Category != "" {
		r.index[tool.Category] = removeFromSlice(r.index[tool.Category], toolID)
	}

	// Remove from tag indices
	for _, tag := range tool.Tags {
		r.tagIdx[tag] = removeFromSlice(r.tagIdx[tag], toolID)
	}

	return nil
}

// GetTool retrieves a tool by ID
func (r *Registry) GetTool(toolID string) (*ToolDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[toolID]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolID)
	}

	return tool, nil
}

// FindByName finds tools by name (case-insensitive partial match)
func (r *Registry) FindByName(name string) []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nameLower := strings.ToLower(name)
	var results []*ToolDefinition

	for _, tool := range r.tools {
		if strings.Contains(strings.ToLower(tool.Name), nameLower) {
			results = append(results, tool)
		}
	}

	return results
}

// FindByCategory finds all tools in a category
func (r *Registry) FindByCategory(category string) []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	toolIDs, exists := r.index[category]
	if !exists {
		return []*ToolDefinition{}
	}

	var results []*ToolDefinition
	for _, id := range toolIDs {
		if tool, ok := r.tools[id]; ok {
			results = append(results, tool)
		}
	}

	return results
}

// FindByTag finds all tools with a tag
func (r *Registry) FindByTag(tag string) []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	toolIDs, exists := r.tagIdx[tag]
	if !exists {
		return []*ToolDefinition{}
	}

	var results []*ToolDefinition
	for _, id := range toolIDs {
		if tool, ok := r.tools[id]; ok {
			results = append(results, tool)
		}
	}

	return results
}

// Search finds tools using multiple criteria
func (r *Registry) Search(query string, category string, tags []string) []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*ToolDefinition
	queryLower := strings.ToLower(query)

	for _, tool := range r.tools {
		// Check category
		if category != "" && tool.Category != category {
			continue
		}

		// Check tags
		if len(tags) > 0 {
			hasTag := false
			for _, tag := range tags {
				for _, toolTag := range tool.Tags {
					if toolTag == tag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Check query matches name or description
		if query != "" {
			if !strings.Contains(strings.ToLower(tool.Name), queryLower) &&
				!strings.Contains(strings.ToLower(tool.Description), queryLower) {
				continue
			}
		}

		results = append(results, tool)
	}

	// Sort by usage count and success rate
	sort.Slice(results, func(i, j int) bool {
		if results[i].SuccessRate != results[j].SuccessRate {
			return results[i].SuccessRate > results[j].SuccessRate
		}
		return results[i].UsageCount > results[j].UsageCount
	})

	return results
}

// RecordUsage records a tool usage
func (r *Registry) RecordUsage(toolID string, agentID string, success bool, latencyMs float64, err string) (*ToolUsage, error) {
	r.mu.Lock()
	tool, exists := r.tools[toolID]
	r.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolID)
	}

	usage := &ToolUsage{
		ID:        uuid.New().String(),
		ToolID:    toolID,
		AgentID:   agentID,
		Success:   success,
		LatencyMs: latencyMs,
		Timestamp: time.Now(),
		Error:     err,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.usage[usage.ID] = usage

	// Update tool statistics
	tool.UsageCount++
	tool.AvgLatencyMs = (tool.AvgLatencyMs*float64(tool.UsageCount-1) + latencyMs) / float64(tool.UsageCount)

	successCount := 0
	for _, u := range r.usage {
		if u.ToolID == toolID && u.Success {
			successCount++
		}
	}
	tool.SuccessRate = float64(successCount) / float64(tool.UsageCount)
	tool.UpdatedAt = time.Now()

	return usage, nil
}

// GetStats gets statistics for a tool
func (r *Registry) GetStats(toolID string) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[toolID]
	if !exists {
		return nil
	}

	// Calculate usage stats
	var totalLatency float64
	successCount := 0
	failureCount := 0

	for _, usage := range r.usage {
		if usage.ToolID == toolID {
			totalLatency += usage.LatencyMs
			if usage.Success {
				successCount++
			} else {
				failureCount++
			}
		}
	}

	return map[string]interface{}{
		"tool_id":       toolID,
		"name":          tool.Name,
		"usage_count":   tool.UsageCount,
		"success_count": successCount,
		"failure_count": failureCount,
		"success_rate":  tool.SuccessRate,
		"avg_latency":   tool.AvgLatencyMs,
		"status":        tool.Status,
		"last_updated":  tool.UpdatedAt,
	}
}

// GetCategories gets all tool categories
func (r *Registry) GetCategories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categoriesMap := make(map[string]bool)
	for cat := range r.index {
		categoriesMap[cat] = true
	}

	var categories []string
	for cat := range categoriesMap {
		categories = append(categories, cat)
	}

	sort.Strings(categories)
	return categories
}

// ListAll lists all tools
func (r *Registry) ListAll() []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tools []*ToolDefinition
	for _, tool := range r.tools {
		if tool.Status != "deprecated" {
			tools = append(tools, tool)
		}
	}

	return tools
}

// GetTopTools gets the most used and reliable tools
func (r *Registry) GetTopTools(limit int) []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tools []*ToolDefinition
	for _, tool := range r.tools {
		if tool.Status != "deprecated" {
			tools = append(tools, tool)
		}
	}

	// Sort by success rate then usage
	sort.Slice(tools, func(i, j int) bool {
		if tools[i].SuccessRate != tools[j].SuccessRate {
			return tools[i].SuccessRate > tools[j].SuccessRate
		}
		return tools[i].UsageCount > tools[j].UsageCount
	})

	if len(tools) > limit {
		tools = tools[:limit]
	}

	return tools
}

// Helper function
func removeFromSlice(slice []string, item string) []string {
	var result []string
	for _, v := range slice {
		if v != item {
			result = append(result, v)
		}
	}
	return result
}
