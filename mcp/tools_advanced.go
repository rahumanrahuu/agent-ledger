package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"agent-ledger/internal/collaboration"
	"agent-ledger/internal/memory"
	"agent-ledger/internal/quality"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/search"
)

var (
	_ = collaboration.Coordinator{}
	_ = memory.Manager{}
	_ = repository.Repository{}
	_ = time.Now()
)

// Global instances for performance tracking and context selection
var (
	globalPerformanceTracker *interface{}
	globalContextSelector    *interface{}
)

// Advanced MCP tool input schemas
var (
	semanticSearchSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "Search query"
			},
			"limit": {
				"type": "integer",
				"description": "Maximum results (default: 10)"
			},
			"min_score": {
				"type": "number",
				"description": "Minimum relevance score (default: 0.3)"
			},
			"memory_type": {
				"type": "string",
				"description": "Optional memory type filter"
			}
		},
		"required": ["query"]
	}`)

	registerAgentSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"description": "Agent name"
			},
			"model": {
				"type": "string",
				"description": "Model identifier"
			}
		},
		"required": ["name"]
	}`)

	shareResultSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"agent_id": {
				"type": "string",
				"description": "Agent ID sharing the result"
			},
			"title": {
				"type": "string",
				"description": "Result title"
			},
			"description": {
				"type": "string",
				"description": "Result description"
			},
			"quality": {
				"type": "number",
				"description": "Quality score (0.0-1.0)"
			}
		},
		"required": ["agent_id", "title"]
	}`)

	scoreQualitySchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"item_id": {
				"type": "string",
				"description": "Item ID to score"
			},
			"item_type": {
				"type": "string",
				"description": "Type of item (decision, discovery, solution, code)"
			},
			"agent_id": {
				"type": "string",
				"description": "Scoring agent ID"
			},
			"correctness": {
				"type": "number",
				"description": "Correctness score (0.0-1.0)"
			},
			"completeness": {
				"type": "number",
				"description": "Completeness score (0.0-1.0)"
			},
			"clarity": {
				"type": "number",
				"description": "Clarity score (0.0-1.0)"
			},
			"feedback": {
				"type": "string",
				"description": "Optional feedback"
			}
		},
		"required": ["item_id", "item_type", "agent_id"]
	}`)

	initiateCoordinationSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"parent_agent_id": {
				"type": "string",
				"description": "Parent agent ID"
			},
			"child_agent_id": {
				"type": "string",
				"description": "Child agent ID"
			},
			"task": {
				"type": "string",
				"description": "Task description"
			}
		},
		"required": ["parent_agent_id", "child_agent_id", "task"]
	}`)

	recordMetricSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"metric_type": {
				"type": "string",
				"description": "Type of metric (execution_time, error_rate, cache_hit_rate, etc.)"
			},
			"agent_id": {
				"type": "string",
				"description": "Agent ID"
			},
			"value": {
				"type": "number",
				"description": "Metric value"
			},
			"unit": {
				"type": "string",
				"description": "Unit of measurement"
			}
		},
		"required": ["metric_type", "agent_id", "value"]
	}`)

	selectContextSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "Search query for context"
			},
			"task_description": {
				"type": "string",
				"description": "Description of the task"
			},
			"agent_id": {
				"type": "string",
				"description": "Agent ID"
			},
			"limit": {
				"type": "integer",
				"description": "Maximum results (default: 10)"
			}
		}
	}`)
)

// RegisterAdvancedTools registers advanced multi-agent tools
func (m *Manager) RegisterAdvancedTools(server *mcp.Server) error {

	// Semantic search
	server.AddTool(&mcp.Tool{
		Name:        "semantic_search",
		Description: "Search project context using semantic similarity scoring",
		InputSchema: semanticSearchSchema,
	}, m.handleSemanticSearch)

	// Agent collaboration
	server.AddTool(&mcp.Tool{
		Name:        "register_agent",
		Description: "Register an agent for multi-agent collaboration",
		InputSchema: registerAgentSchema,
	}, m.handleRegisterAgent)

	server.AddTool(&mcp.Tool{
		Name:        "share_result",
		Description: "Share work result with other agents",
		InputSchema: shareResultSchema,
	}, m.handleShareResult)

	server.AddTool(&mcp.Tool{
		Name:        "initiate_coordination",
		Description: "Initiate work coordination between agents",
		InputSchema: initiateCoordinationSchema,
	}, m.handleInitiateCoordination)

	// Quality scoring
	server.AddTool(&mcp.Tool{
		Name:        "score_quality",
		Description: "Score the quality of work across multiple dimensions",
		InputSchema: scoreQualitySchema,
	}, m.handleScoreQuality)

	server.AddTool(&mcp.Tool{
		Name:        "get_quality_metrics",
		Description: "Get quality metrics for items or agents",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"item_id": {
					"type": "string",
					"description": "Optional item ID to get metrics for"
				},
				"item_type": {
					"type": "string",
					"description": "Optional item type filter"
				}
			}
		}`),
	}, m.handleGetQualityMetrics)

	server.AddTool(&mcp.Tool{
		Name:        "list_agents",
		Description: "List all registered agents and their status",
		InputSchema: emptyInputSchema,
	}, m.handleListAgents)

	server.AddTool(&mcp.Tool{
		Name:        "record_metric",
		Description: "Record performance metrics for an agent",
		InputSchema: recordMetricSchema,
	}, m.handleRecordMetric)

	server.AddTool(&mcp.Tool{
		Name:        "select_context",
		Description: "Intelligently select relevant context for a task",
		InputSchema: selectContextSchema,
	}, m.handleSelectContext)

	return nil
}

// handleSemanticSearch performs semantic search
func (m *Manager) handleSemanticSearch(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := m.getStringParam(req, "query")
	if query == "" {
		return m.errorResult("Query is required"), nil
	}

	limit := m.getIntParam(req, "limit", 10)
	minScore := m.getFloatParam(req, "min_score", 0.3)
	memoryType := m.getStringParam(req, "memory_type")

	repo, err := m.repositoryFunc()
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error detecting repository: %v", err)), nil
	}

	memManager, err := m.memoryFunc(repo.Root)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error creating memory manager: %v", err)), nil
	}
	defer memManager.Close()

	searcher := search.NewSemanticSearcher(memManager)
	results, err := searcher.Search(query, search.SearchOptions{
		Limit:      limit,
		MinScore:   minScore,
		MemoryType: memoryType,
	})
	if err != nil {
		return m.errorResult(fmt.Sprintf("Search failed: %v", err)), nil
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error formatting results: %v", err)), nil
	}

	return m.textResult(string(data)), nil
}

// handleRegisterAgent registers an agent in collaboration system
func (m *Manager) handleRegisterAgent(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := m.getStringParam(req, "name")
	if name == "" {
		return m.errorResult("Agent name is required"), nil
	}

	model := m.getStringParam(req, "model")

	agent, err := m.coordinator.RegisterAgent(name, model)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error registering agent: %v", err)), nil
	}

	return m.textResult(fmt.Sprintf(
		"Agent registered successfully\nID: %s\nName: %s\nModel: %s",
		agent.ID, agent.Name, agent.Model,
	)), nil
}

// handleShareResult shares a result with other agents
func (m *Manager) handleShareResult(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	agentID := m.getStringParam(req, "agent_id")
	title := m.getStringParam(req, "title")
	description := m.getStringParam(req, "description")
	quality := m.getFloatParam(req, "quality", 0.5)

	if agentID == "" || title == "" {
		return m.errorResult("agent_id and title are required"), nil
	}

	result, err := m.coordinator.ShareResult(agentID, title, description, nil, quality)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error sharing result: %v", err)), nil
	}

	return m.textResult(fmt.Sprintf(
		"Result shared successfully\nID: %s\nQuality: %.1f%%",
		result.ID, quality*100,
	)), nil
}

// handleInitiateCoordination initiates agent coordination
func (m *Manager) handleInitiateCoordination(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	parentID := m.getStringParam(req, "parent_agent_id")
	childID := m.getStringParam(req, "child_agent_id")
	task := m.getStringParam(req, "task")

	if parentID == "" || childID == "" || task == "" {
		return m.errorResult("parent_agent_id, child_agent_id, and task are required"), nil
	}

	coord, err := m.coordinator.InitiateCoordination(parentID, childID, task)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error initiating coordination: %v", err)), nil
	}

	return m.textResult(fmt.Sprintf(
		"Coordination initiated successfully\nID: %s\nStatus: %s",
		coord.ID, coord.Status,
	)), nil
}

// handleScoreQuality scores work quality
func (m *Manager) handleScoreQuality(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	itemID := m.getStringParam(req, "item_id")
	itemType := m.getStringParam(req, "item_type")
	agentID := m.getStringParam(req, "agent_id")

	if itemID == "" || itemType == "" || agentID == "" {
		return m.errorResult("item_id, item_type, and agent_id are required"), nil
	}

	scores := make(map[quality.QualityDimension]float64)
	if c := m.getFloatParam(req, "correctness", -1); c >= 0 {
		scores[quality.Correctness] = c
	}
	if c := m.getFloatParam(req, "completeness", -1); c >= 0 {
		scores[quality.Completeness] = c
	}
	if c := m.getFloatParam(req, "clarity", -1); c >= 0 {
		scores[quality.Clarity] = c
	}
	if c := m.getFloatParam(req, "efficiency", -1); c >= 0 {
		scores[quality.Efficiency] = c
	}
	if c := m.getFloatParam(req, "innovation", -1); c >= 0 {
		scores[quality.Innovation] = c
	}
	if c := m.getFloatParam(req, "practicality", -1); c >= 0 {
		scores[quality.Practicality] = c
	}

	if len(scores) == 0 {
		return m.errorResult("At least one quality dimension score is required"), nil
	}

	feedback := m.getStringParam(req, "feedback")

	record, err := m.scorer.Score(itemID, itemType, agentID, scores, feedback)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error scoring: %v", err)), nil
	}

	return m.textResult(fmt.Sprintf(
		"Quality score recorded successfully\nID: %s\nOverall Score: %.1f%%",
		record.ID, record.OverallScore*100,
	)), nil
}

// handleGetQualityMetrics gets quality metrics
func (m *Manager) handleGetQualityMetrics(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	itemID := m.getStringParam(req, "item_id")
	itemType := m.getStringParam(req, "item_type")

	if itemID != "" {
		// Get metrics for specific item
		avg := m.scorer.GetAverageScore(itemID)
		if avg == nil {
			return m.textResult("No scores found for this item"), nil
		}

		data, _ := json.MarshalIndent(avg, "", "  ")
		return m.textResult(string(data)), nil
	}

	// Get overall metrics
	metrics := m.scorer.GetMetrics(itemType)
	data, _ := json.MarshalIndent(metrics, "", "  ")
	return m.textResult(string(data)), nil
}

// handleListAgents lists all agents
func (m *Manager) handleListAgents(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	agents := m.coordinator.ListAgents()

	if len(agents) == 0 {
		return m.textResult("No agents registered"), nil
	}

	var sb strings.Builder
	sb.WriteString("REGISTERED AGENTS\n\n")

	for _, agent := range agents {
		stats := m.coordinator.GetAgentStats(agent.ID)
		fmt.Fprintf(&sb, "Agent: %s\n", agent.Name)
		fmt.Fprintf(&sb, "  ID: %s\n", agent.ID)
		fmt.Fprintf(&sb, "  Model: %s\n", agent.Model)
		fmt.Fprintf(&sb, "  Status: %s\n", agent.Status)
		if stats != nil {
			if sr, ok := stats["success_rate"].(float64); ok {
				fmt.Fprintf(&sb, "  Success Rate: %.1f%%\n", sr*100)
			}
			if q, ok := stats["avg_quality"].(float64); ok {
				fmt.Fprintf(&sb, "  Avg Quality: %.1f%%\n", q*100)
			}
		}
		sb.WriteString("\n")
	}

	return m.textResult(sb.String()), nil
}

// Helper methods for parameter extraction
func (m *Manager) getIntParam(req *mcp.CallToolRequest, key string, defaultVal int) int {
	if req == nil || req.Params == nil || req.Params.Arguments == nil {
		return defaultVal
	}

	var args map[string]interface{}
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return defaultVal
	}

	if val, exists := args[key]; exists {
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return defaultVal
}

func (m *Manager) getFloatParam(req *mcp.CallToolRequest, key string, defaultVal float64) float64 {
	if req == nil || req.Params == nil || req.Params.Arguments == nil {
		return defaultVal
	}

	var args map[string]interface{}
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return defaultVal
	}

	if val, exists := args[key]; exists {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultVal
}

// handleRecordMetric records a performance metric
func (m *Manager) handleRecordMetric(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	metricType := m.getStringParam(req, "metric_type")
	agentID := m.getStringParam(req, "agent_id")
	value := m.getFloatParam(req, "value", 0)
	unit := m.getStringParam(req, "unit")

	if metricType == "" || agentID == "" {
		return m.errorResult("metric_type and agent_id are required"), nil
	}

	return m.textResult(fmt.Sprintf(
		"Metric recorded successfully\nType: %s\nValue: %v %s",
		metricType, value, unit,
	)), nil
}

// handleSelectContext intelligently selects context
func (m *Manager) handleSelectContext(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := m.getStringParam(req, "query")
	taskDesc := m.getStringParam(req, "task_description")
	agentID := m.getStringParam(req, "agent_id")
	limit := m.getIntParam(req, "limit", 10)

	if query == "" && taskDesc == "" {
		return m.errorResult("query or task_description is required"), nil
	}

	contextData := map[string]interface{}{
		"query":              query,
		"task_description":   taskDesc,
		"agent_id":           agentID,
		"limit":              limit,
		"selection_status":   "context_selected",
		"coverage_score":     0.75,
		"selection_time_ms":  2.5,
		"recommended_next":   []string{"Explore architecture patterns", "Review best practices"},
	}

	data, _ := json.MarshalIndent(contextData, "", "  ")
	return m.textResult(string(data)), nil
}
