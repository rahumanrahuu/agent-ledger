package analytics

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MetricType represents a type of performance metric
type MetricType string

const (
	ExecutionTime   MetricType = "execution_time"
	MemoryUsage     MetricType = "memory_usage"
	CacheHitRate    MetricType = "cache_hit_rate"
	ErrorRate       MetricType = "error_rate"
	SuccessRate     MetricType = "success_rate"
	ThroughputRate  MetricType = "throughput_rate"
	LatencyP50      MetricType = "latency_p50"
	LatencyP95      MetricType = "latency_p95"
	LatencyP99      MetricType = "latency_p99"
)

// Metric represents a performance metric data point
type Metric struct {
	ID        string    `json:"id"`
	Type      MetricType `json:"type"`
	AgentID   string    `json:"agent_id"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
	Labels    map[string]string `json:"labels"`
}

// PerformanceTracker tracks agent performance metrics
type PerformanceTracker struct {
	mu      sync.RWMutex
	metrics map[string]*Metric
	index   map[string][]string // agentID -> metricIDs
	windows map[string]*MetricWindow
}

// MetricWindow aggregates metrics over time
type MetricWindow struct {
	AgentID   string    `json:"agent_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Metrics   map[MetricType]*AggregatedMetric `json:"metrics"`
}

// AggregatedMetric represents aggregated metric data
type AggregatedMetric struct {
	Type   MetricType `json:"type"`
	Count  int        `json:"count"`
	Sum    float64    `json:"sum"`
	Mean   float64    `json:"mean"`
	Min    float64    `json:"min"`
	Max    float64    `json:"max"`
	StdDev float64    `json:"std_dev"`
	P50    float64    `json:"p50"`
	P95    float64    `json:"p95"`
	P99    float64    `json:"p99"`
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		metrics: make(map[string]*Metric),
		index:   make(map[string][]string),
		windows: make(map[string]*MetricWindow),
	}
}

// RecordMetric records a performance metric
func (pt *PerformanceTracker) RecordMetric(metricType MetricType, agentID string, value float64, unit string) (*Metric, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agentID is required")
	}

	metric := &Metric{
		ID:        uuid.New().String(),
		Type:      metricType,
		AgentID:   agentID,
		Value:     value,
		Unit:      unit,
		Timestamp: time.Now(),
		Labels:    make(map[string]string),
	}

	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.metrics[metric.ID] = metric
	pt.index[agentID] = append(pt.index[agentID], metric.ID)

	return metric, nil
}

// GetMetrics retrieves metrics for an agent
func (pt *PerformanceTracker) GetMetrics(agentID string, metricType MetricType, limit int) []*Metric {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	metricIDs, exists := pt.index[agentID]
	if !exists {
		return []*Metric{}
	}

	var metrics []*Metric
	for _, id := range metricIDs {
		if metric, ok := pt.metrics[id]; ok {
			if metricType == "" || metric.Type == metricType {
				metrics = append(metrics, metric)
			}
		}
	}

	// Sort by timestamp descending
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Timestamp.After(metrics[j].Timestamp)
	})

	if len(metrics) > limit && limit > 0 {
		metrics = metrics[:limit]
	}

	return metrics
}

// AnalyzeWindow creates a metric window for analysis
func (pt *PerformanceTracker) AnalyzeWindow(agentID string, duration time.Duration) *MetricWindow {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	metricIDs, exists := pt.index[agentID]
	if !exists {
		return nil
	}

	cutoff := time.Now().Add(-duration)
	window := &MetricWindow{
		AgentID:   agentID,
		StartTime: cutoff,
		EndTime:   time.Now(),
		Metrics:   make(map[MetricType]*AggregatedMetric),
	}

	// Collect metrics by type
	metricsMap := make(map[MetricType][]float64)

	for _, id := range metricIDs {
		if metric, ok := pt.metrics[id]; ok && metric.Timestamp.After(cutoff) {
			metricsMap[metric.Type] = append(metricsMap[metric.Type], metric.Value)
		}
	}

	// Aggregate each metric type
	for metricType, values := range metricsMap {
		if len(values) > 0 {
			window.Metrics[metricType] = aggregateValues(metricType, values)
		}
	}

	return window
}

// aggregateValues computes statistics for a set of values
func aggregateValues(metricType MetricType, values []float64) *AggregatedMetric {
	agg := &AggregatedMetric{
		Type:  metricType,
		Count: len(values),
	}

	if len(values) == 0 {
		return agg
	}

	// Sort for percentile calculations
	sort.Float64s(values)

	// Calculate sum and min/max
	sum := 0.0
	minVal := values[0]
	maxVal := values[0]

	for _, v := range values {
		sum += v
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	agg.Sum = sum
	agg.Mean = sum / float64(len(values))
	agg.Min = minVal
	agg.Max = maxVal

	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		variance += (v - agg.Mean) * (v - agg.Mean)
	}
	agg.StdDev = variance / float64(len(values))

	// Calculate percentiles
	agg.P50 = getPercentile(values, 0.50)
	agg.P95 = getPercentile(values, 0.95)
	agg.P99 = getPercentile(values, 0.99)

	return agg
}

// getPercentile calculates the nth percentile
func getPercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	index := int(float64(len(values)-1) * percentile)
	return values[index]
}

// CompareAgents compares performance across agents
func (pt *PerformanceTracker) CompareAgents(agentIDs []string, metricType MetricType, duration time.Duration) map[string]*AggregatedMetric {
	result := make(map[string]*AggregatedMetric)

	for _, agentID := range agentIDs {
		window := pt.AnalyzeWindow(agentID, duration)
		if window != nil && window.Metrics[metricType] != nil {
			result[agentID] = window.Metrics[metricType]
		}
	}

	return result
}

// GetTrendAnalysis analyzes metric trends over time
func (pt *PerformanceTracker) GetTrendAnalysis(agentID string, metricType MetricType, periods int) []TrendPoint {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	metricIDs, exists := pt.index[agentID]
	if !exists || periods <= 0 {
		return []TrendPoint{}
	}

	// Group metrics by period
	now := time.Now()
	periodDuration := (24 * time.Hour) / time.Duration(periods)
	periodMetrics := make([][]float64, periods)

	for _, id := range metricIDs {
		if metric, ok := pt.metrics[id]; ok && metric.Type == metricType {
			age := now.Sub(metric.Timestamp)
			periodIndex := int(age / periodDuration)
			if periodIndex >= 0 && periodIndex < periods {
				periodMetrics[periodIndex] = append(periodMetrics[periodIndex], metric.Value)
			}
		}
	}

	// Calculate trend points
	var trends []TrendPoint
	for i, values := range periodMetrics {
		if len(values) > 0 {
			var sum float64
			for _, v := range values {
				sum += v
			}
			trends = append(trends, TrendPoint{
				Period:    i,
				Value:     sum / float64(len(values)),
				Count:     len(values),
				Timestamp: now.Add(-time.Duration(i) * periodDuration),
			})
		}
	}

	return trends
}

// TrendPoint represents a point in a trend analysis
type TrendPoint struct {
	Period    int       `json:"period"`
	Value     float64   `json:"value"`
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthCheck performs a health check on agent performance
func (pt *PerformanceTracker) HealthCheck(agentID string) map[string]interface{} {
	window := pt.AnalyzeWindow(agentID, 1*time.Hour)
	if window == nil {
		return map[string]interface{}{
			"status": "unknown",
			"reason": "no recent metrics",
		}
	}

	status := "healthy"
	issues := []string{}

	// Check error rate
	if errMetric, ok := window.Metrics[ErrorRate]; ok {
		if errMetric.Mean > 0.1 {
			status = "warning"
			issues = append(issues, fmt.Sprintf("high error rate: %.1f%%", errMetric.Mean*100))
		}
	}

	// Check latency
	if latencyMetric, ok := window.Metrics[LatencyP99]; ok {
		if latencyMetric.Mean > 5000 {
			status = "warning"
			issues = append(issues, fmt.Sprintf("high latency p99: %.0fms", latencyMetric.Mean))
		}
	}

	// Check cache hit rate
	if cacheMetric, ok := window.Metrics[CacheHitRate]; ok {
		if cacheMetric.Mean < 0.5 {
			issues = append(issues, fmt.Sprintf("low cache hit rate: %.1f%%", cacheMetric.Mean*100))
		}
	}

	if len(issues) > 0 && status != "error" {
		status = "warning"
	}

	return map[string]interface{}{
		"agent_id":  agentID,
		"status":    status,
		"issues":    issues,
		"metrics":   window.Metrics,
		"checked_at": time.Now(),
	}
}
