package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// TraceType represents the type of trace
type TraceType string

const (
	MemoryRetrievalTrace  TraceType = "memory_retrieval"
	ToolCallTrace         TraceType = "tool_call"
	DecisionPointTrace    TraceType = "decision_point"
	ConstraintCheckTrace  TraceType = "constraint_check"
	ErrorPointTrace       TraceType = "error_point"
)

// Trace represents a single reasoning trace
type Trace struct {
	Type         TraceType              `json:"type"`
	Timestamp    time.Time              `json:"timestamp"`
	Label        string                 `json:"label"`
	Duration     int                    `json:"duration_ms"`
	Score        float64                `json:"score,omitempty"`
	Success      bool                   `json:"success"`
	Details      map[string]interface{} `json:"details"`
}

// TraceRecorder records reasoning traces for debugging and analysis
type TraceRecorder struct {
	sessionID string
	root      string
	traces    []Trace
}

// NewTraceRecorder creates a new trace recorder
func NewTraceRecorder(root, sessionID string) *TraceRecorder {
	return &TraceRecorder{
		sessionID: sessionID,
		root:      root,
		traces:    []Trace{},
	}
}

// RecordTrace records a trace
func (tr *TraceRecorder) RecordTrace(t Trace) {
	if t.Timestamp.IsZero() {
		t.Timestamp = time.Now()
	}
	tr.traces = append(tr.traces, t)
}

// RecordMemoryRetrieval records a memory retrieval trace
func (tr *TraceRecorder) RecordMemoryRetrieval(query string, count int, score float64, latencyMs int) {
	tr.RecordTrace(Trace{
		Type:      MemoryRetrievalTrace,
		Timestamp: time.Now(),
		Label:     "Retrieved: " + query,
		Duration:  latencyMs,
		Score:     score,
		Success:   count > 0,
		Details: map[string]interface{}{
			"query":         query,
			"results_count": count,
		},
	})
}

// RecordToolCall records a tool call trace
func (tr *TraceRecorder) RecordToolCall(toolName string, success bool, latencyMs int, costTokens int) {
	tr.RecordTrace(Trace{
		Type:      ToolCallTrace,
		Timestamp: time.Now(),
		Label:     "Tool: " + toolName,
		Duration:  latencyMs,
		Success:   success,
		Details: map[string]interface{}{
			"tool":        toolName,
			"cost_tokens": costTokens,
		},
	})
}

// RecordDecisionPoint records a decision point
func (tr *TraceRecorder) RecordDecisionPoint(question string, chosen string, reasoning string) {
	tr.RecordTrace(Trace{
		Type:      DecisionPointTrace,
		Timestamp: time.Now(),
		Label:     question,
		Success:   true,
		Details: map[string]interface{}{
			"chosen":    chosen,
			"reasoning": reasoning,
		},
	})
}

// RecordConstraintCheck records a constraint check
func (tr *TraceRecorder) RecordConstraintCheck(constraintID string, passed bool, details string) {
	tr.RecordTrace(Trace{
		Type:      ConstraintCheckTrace,
		Timestamp: time.Now(),
		Label:     "Constraint: " + constraintID,
		Success:   passed,
		Details: map[string]interface{}{
			"constraint": constraintID,
			"details":    details,
		},
	})
}

// Export exports traces to file
func (tr *TraceRecorder) Export() error {
	traceDir := filepath.Join(tr.root, ".agent", "events", tr.sessionID, "traces")
	if err := os.MkdirAll(traceDir, 0755); err != nil {
		return err
	}

	filename := time.Now().Format("20060102-150405") + ".json"
	filepath := filepath.Join(traceDir, filename)

	data, err := json.MarshalIndent(tr.traces, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// GenerateSummary generates a summary of traces
type TraceSummary struct {
	TotalTraces        int     `json:"total_traces"`
	MemoryRetrievals   int     `json:"memory_retrievals"`
	ToolCalls          int     `json:"tool_calls"`
	DecisionPoints     int     `json:"decision_points"`
	ConstraintChecks   int     `json:"constraint_checks"`
	AverageScore       float64 `json:"average_score"`
	TotalLatency       int     `json:"total_latency_ms"`
	SuccessRate        float64 `json:"success_rate"`
}

// GenerateSummary generates a trace summary
func (tr *TraceRecorder) GenerateSummary() TraceSummary {
	summary := TraceSummary{
		TotalTraces: len(tr.traces),
	}

	var scoreSum float64
	var successCount int
	var totalLatency int

	for _, t := range tr.traces {
		switch t.Type {
		case MemoryRetrievalTrace:
			summary.MemoryRetrievals++
		case ToolCallTrace:
			summary.ToolCalls++
		case DecisionPointTrace:
			summary.DecisionPoints++
		case ConstraintCheckTrace:
			summary.ConstraintChecks++
		}

		if t.Success {
			successCount++
		}
		scoreSum += t.Score
		totalLatency += t.Duration
	}

	if summary.TotalTraces > 0 {
		summary.AverageScore = scoreSum / float64(summary.TotalTraces)
		summary.SuccessRate = float64(successCount) / float64(summary.TotalTraces)
	}

	summary.TotalLatency = totalLatency
	return summary
}
