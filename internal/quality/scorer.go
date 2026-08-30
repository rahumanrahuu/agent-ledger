package quality

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// QualityDimension represents a quality assessment dimension
type QualityDimension string

const (
	Correctness    QualityDimension = "correctness"
	Completeness   QualityDimension = "completeness"
	Clarity        QualityDimension = "clarity"
	Efficiency     QualityDimension = "efficiency"
	Innovation     QualityDimension = "innovation"
	Practicality   QualityDimension = "practicality"
	Testability    QualityDimension = "testability"
	Documentation  QualityDimension = "documentation"
)

// ScoreRecord represents a quality score
type ScoreRecord struct {
	ID            string                   `json:"id"`
	ItemID        string                   `json:"item_id"`
	ItemType      string                   `json:"item_type"` // decision, discovery, solution, code
	ScoredByAgent string                   `json:"scored_by_agent"`
	Scores        map[QualityDimension]float64 `json:"scores"`
	OverallScore  float64                  `json:"overall_score"`
	Feedback      string                   `json:"feedback"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

// Scorer rates quality of work
type Scorer struct {
	mu      sync.RWMutex
	records map[string]*ScoreRecord
	index   map[string][]string // itemID -> scoreIDs
}

// NewScorer creates a new quality scorer
func NewScorer() *Scorer {
	return &Scorer{
		records: make(map[string]*ScoreRecord),
		index:   make(map[string][]string),
	}
}

// Score creates a quality assessment
func (s *Scorer) Score(itemID, itemType, agentID string, dimensionScores map[QualityDimension]float64, feedback string) (*ScoreRecord, error) {
	if itemID == "" || itemType == "" {
		return nil, fmt.Errorf("itemID and itemType are required")
	}

	// Validate scores are in range
	for dim, score := range dimensionScores {
		if score < 0 || score > 1 {
			return nil, fmt.Errorf("score for %s must be between 0 and 1", dim)
		}
	}

	// Calculate overall score as weighted average
	overallScore := s.calculateOverallScore(dimensionScores)

	record := &ScoreRecord{
		ID:            uuid.New().String(),
		ItemID:        itemID,
		ItemType:      itemType,
		ScoredByAgent: agentID,
		Scores:        dimensionScores,
		OverallScore:  overallScore,
		Feedback:      feedback,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[record.ID] = record
	s.index[itemID] = append(s.index[itemID], record.ID)

	return record, nil
}

// calculateOverallScore computes weighted overall score
func (s *Scorer) calculateOverallScore(scores map[QualityDimension]float64) float64 {
	weights := map[QualityDimension]float64{
		Correctness:   0.25,
		Completeness:  0.20,
		Clarity:       0.15,
		Efficiency:    0.15,
		Innovation:    0.10,
		Practicality:  0.10,
		Testability:   0.03,
		Documentation: 0.02,
	}

	totalScore := 0.0
	totalWeight := 0.0

	for dim, weight := range weights {
		if score, exists := scores[dim]; exists {
			totalScore += score * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return totalScore / totalWeight
}

// GetScores retrieves all scores for an item
func (s *Scorer) GetScores(itemID string) []*ScoreRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scoreIDs, exists := s.index[itemID]
	if !exists {
		return []*ScoreRecord{}
	}

	var scores []*ScoreRecord
	for _, id := range scoreIDs {
		if record, ok := s.records[id]; ok {
			scores = append(scores, record)
		}
	}

	return scores
}

// GetAverageScore gets the average score for an item
func (s *Scorer) GetAverageScore(itemID string) *AggregateScore {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scoreIDs, exists := s.index[itemID]
	if !exists {
		return nil
	}

	var records []*ScoreRecord
	for _, id := range scoreIDs {
		if record, ok := s.records[id]; ok {
			records = append(records, record)
		}
	}

	if len(records) == 0 {
		return nil
	}

	return aggregateScores(records)
}

// AggregateScore represents aggregated quality scores
type AggregateScore struct {
	ItemID            string                        `json:"item_id"`
	AverageScore      float64                       `json:"average_score"`
	DimensionAverages map[QualityDimension]float64 `json:"dimension_averages"`
	ScoreCount        int                           `json:"score_count"`
	HighestScore      float64                       `json:"highest_score"`
	LowestScore       float64                       `json:"lowest_score"`
	Trend             string                        `json:"trend"` // improving, stable, declining
}

// aggregateScores combines multiple score records
func aggregateScores(records []*ScoreRecord) *AggregateScore {
	if len(records) == 0 {
		return nil
	}

	agg := &AggregateScore{
		ItemID:            records[0].ItemID,
		DimensionAverages: make(map[QualityDimension]float64),
		ScoreCount:        len(records),
		HighestScore:      0,
		LowestScore:       1,
	}

	// Aggregate by dimension
	dimensionTotals := make(map[QualityDimension]float64)
	dimensionCounts := make(map[QualityDimension]int)
	var overallScores []float64

	for _, record := range records {
		overallScores = append(overallScores, record.OverallScore)

		if record.OverallScore > agg.HighestScore {
			agg.HighestScore = record.OverallScore
		}
		if record.OverallScore < agg.LowestScore {
			agg.LowestScore = record.OverallScore
		}

		for dim, score := range record.Scores {
			dimensionTotals[dim] += score
			dimensionCounts[dim]++
		}
	}

	// Calculate dimension averages
	for dim, total := range dimensionTotals {
		count := dimensionCounts[dim]
		if count > 0 {
			agg.DimensionAverages[dim] = total / float64(count)
		}
	}

	// Calculate overall average
	var sum float64
	for _, score := range overallScores {
		sum += score
	}
	agg.AverageScore = sum / float64(len(overallScores))

	// Determine trend
	if len(records) > 2 {
		recent := overallScores[len(overallScores)-1]
		older := overallScores[0]
		diff := recent - older
		if diff > 0.1 {
			agg.Trend = "improving"
		} else if diff < -0.1 {
			agg.Trend = "declining"
		} else {
			agg.Trend = "stable"
		}
	}

	return agg
}

// GetTopItems gets highest-scoring items
func (s *Scorer) GetTopItems(itemType string, limit int) []*AggregateScore {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var aggregates []*AggregateScore

	// Collect all unique items
	itemMap := make(map[string][]*ScoreRecord)
	for _, record := range s.records {
		if itemType == "" || record.ItemType == itemType {
			itemMap[record.ItemID] = append(itemMap[record.ItemID], record)
		}
	}

	// Aggregate scores
	for _, records := range itemMap {
		agg := aggregateScores(records)
		if agg != nil {
			aggregates = append(aggregates, agg)
		}
	}

	// Sort by score descending
	for i := 0; i < len(aggregates)-1; i++ {
		for j := i + 1; j < len(aggregates); j++ {
			if aggregates[j].AverageScore > aggregates[i].AverageScore {
				aggregates[i], aggregates[j] = aggregates[j], aggregates[i]
			}
		}
	}

	if len(aggregates) > limit && limit > 0 {
		aggregates = aggregates[:limit]
	}

	return aggregates
}

// GetMetrics gets quality metrics summary
func (s *Scorer) GetMetrics(itemType string) map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var records []*ScoreRecord
	for _, record := range s.records {
		if itemType == "" || record.ItemType == itemType {
			records = append(records, record)
		}
	}

	if len(records) == 0 {
		return map[string]interface{}{
			"total_scores": 0,
		}
	}

	// Calculate statistics
	var sum, minScore, maxScore float64 = 0, 1, 0
	dimensionStats := make(map[QualityDimension]map[string]float64)

	for _, record := range records {
		sum += record.OverallScore
		if record.OverallScore < minScore {
			minScore = record.OverallScore
		}
		if record.OverallScore > maxScore {
			maxScore = record.OverallScore
		}

		for dim, score := range record.Scores {
			if dimensionStats[dim] == nil {
				dimensionStats[dim] = make(map[string]float64)
			}
			dimensionStats[dim]["sum"] += score
			dimensionStats[dim]["count"]++
		}
	}

	// Calculate dimension averages
	dimensionAverages := make(map[string]float64)
	for dim, stats := range dimensionStats {
		if stats["count"] > 0 {
			dimensionAverages[string(dim)] = stats["sum"] / stats["count"]
		}
	}

	return map[string]interface{}{
		"total_scores":          len(records),
		"average_score":         sum / float64(len(records)),
		"min_score":             minScore,
		"max_score":             maxScore,
		"dimension_averages":    dimensionAverages,
		"score_distribution":    getScoreDistribution(records),
	}
}

// getScoreDistribution categorizes scores into buckets
func getScoreDistribution(records []*ScoreRecord) map[string]int {
	dist := map[string]int{
		"excellent": 0, // 0.9-1.0
		"good":      0, // 0.7-0.89
		"fair":      0, // 0.5-0.69
		"poor":      0, // 0.3-0.49
		"critical":  0, // 0.0-0.29
	}

	for _, record := range records {
		score := record.OverallScore
		if score >= 0.9 {
			dist["excellent"]++
		} else if score >= 0.7 {
			dist["good"]++
		} else if score >= 0.5 {
			dist["fair"]++
		} else if score >= 0.3 {
			dist["poor"]++
		} else {
			dist["critical"]++
		}
	}

	return dist
}

// RecommendImprovements suggests improvements based on scores
func (s *Scorer) RecommendImprovements(itemID string) []string {
	avg := s.GetAverageScore(itemID)
	if avg == nil {
		return []string{}
	}

	var recommendations []string

	// Check dimensions below average
	dimensions := []QualityDimension{
		Correctness, Completeness, Clarity, Efficiency,
		Innovation, Practicality, Testability, Documentation,
	}

	overallAvg := avg.AverageScore
	for _, dim := range dimensions {
		if score, exists := avg.DimensionAverages[dim]; exists {
			if score < overallAvg-0.15 {
				recommendations = append(
					recommendations,
					fmt.Sprintf("Improve %s (score: %.1f%%)", strings.ToLower(string(dim)), score*100),
				)
			}
		}
	}

	return recommendations
}
