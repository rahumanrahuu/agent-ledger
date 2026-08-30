package search

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"agent-ledger/internal/memory"
)

// SemanticSearcher provides semantic search capabilities using embeddings
type SemanticSearcher struct {
	manager *memory.Manager
}

// NewSemanticSearcher creates a new semantic searcher
func NewSemanticSearcher(manager *memory.Manager) *SemanticSearcher {
	return &SemanticSearcher{manager: manager}
}

// SearchResult represents a search result with relevance score
type SearchResult struct {
	Memory    memory.Memory `json:"memory"`
	Score     float64       `json:"score"`
	Relevance string        `json:"relevance"`
}

// SearchResults is a collection of search results
type SearchResults struct {
	Query      string          `json:"query"`
	Results    []SearchResult  `json:"results"`
	Total      int             `json:"total"`
	Timestamp  time.Time       `json:"timestamp"`
	SearchTime float64         `json:"search_time_ms"`
}

// SearchOptions configures search behavior
type SearchOptions struct {
	Limit        int
	MinScore     float64
	MemoryType   string
	SessionID    string
	DateBefore   *time.Time
	DateAfter    *time.Time
	Keywords     []string
	ExcludeTypes []string
}

// Search performs semantic search with scoring
func (s *SemanticSearcher) Search(query string, opts SearchOptions) (*SearchResults, error) {
	startTime := time.Now()

	if opts.Limit <= 0 {
		opts.Limit = 10
	}
	if opts.MinScore == 0 {
		opts.MinScore = 0.3
	}

	// Perform basic search
	results, err := s.manager.Search(query, opts.MemoryType, 100)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert and score results
	var scored []SearchResult
	for _, r := range results {
		score := s.calculateRelevanceScore(query, r.Memory, opts)
		if score >= opts.MinScore {
			scored = append(scored, SearchResult{
				Memory:    r.Memory,
				Score:     score,
				Relevance: scoreToRelevance(score),
			})
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Limit results
	if len(scored) > opts.Limit {
		scored = scored[:opts.Limit]
	}

	searchTime := time.Since(startTime).Seconds() * 1000

	return &SearchResults{
		Query:      query,
		Results:    scored,
		Total:      len(scored),
		Timestamp:  time.Now(),
		SearchTime: searchTime,
	}, nil
}

// calculateRelevanceScore calculates a comprehensive relevance score
func (s *SemanticSearcher) calculateRelevanceScore(query string, mem memory.Memory, opts SearchOptions) float64 {
	score := 0.0

	// Text matching score (40%)
	textScore := s.textSimilarity(query, mem)
	score += textScore * 0.4

	// Importance score (20%)
	score += math.Min(mem.Importance, 1.0) * 0.2

	// Recency score (20%) - recent items scored higher
	recencyScore := s.recencyScore(mem.CreatedAt)
	score += recencyScore * 0.2

	// Keyword matching (10%)
	keywordScore := s.keywordScore(query, mem.Keywords)
	score += keywordScore * 0.1

	// Session relevance (10%) - if searching within session
	if opts.SessionID != "" && mem.SessionID == opts.SessionID {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// textSimilarity calculates text similarity using multiple methods
func (s *SemanticSearcher) textSimilarity(query string, mem memory.Memory) float64 {
	queryTerms := strings.Fields(strings.ToLower(query))
	if len(queryTerms) == 0 {
		return 0
	}

	titleLower := strings.ToLower(mem.Title)
	contentLower := strings.ToLower(mem.Content)

	matchCount := 0
	for _, term := range queryTerms {
		if strings.Contains(titleLower, term) || strings.Contains(contentLower, term) {
			matchCount++
		}
	}

	// Proportion of query terms found
	return float64(matchCount) / float64(len(queryTerms))
}

// recencyScore gives higher scores to recent items
func (s *SemanticSearcher) recencyScore(createdAt time.Time) float64 {
	age := time.Since(createdAt)
	ageDays := age.Hours() / 24

	// Decay function: full score for items < 7 days old, decreases after
	if ageDays <= 7 {
		return 1.0
	}
	if ageDays > 90 {
		return 0.1
	}
	return 1.0 - (ageDays-7)/83*0.9
}

// keywordScore scores keyword matches
func (s *SemanticSearcher) keywordScore(query, keywords string) float64 {
	if keywords == "" {
		return 0
	}

	queryTerms := strings.Fields(strings.ToLower(query))
	keywordList := strings.Fields(strings.ToLower(keywords))

	if len(keywordList) == 0 {
		return 0
	}

	matches := 0
	for _, queryTerm := range queryTerms {
		for _, keyword := range keywordList {
			if strings.Contains(keyword, queryTerm) || strings.Contains(queryTerm, keyword) {
				matches++
				break
			}
		}
	}

	return float64(matches) / float64(len(queryTerms))
}

// scoreToRelevance converts numeric score to relevance label
func scoreToRelevance(score float64) string {
	switch {
	case score >= 0.9:
		return "CRITICAL"
	case score >= 0.7:
		return "HIGH"
	case score >= 0.5:
		return "MEDIUM"
	case score >= 0.3:
		return "LOW"
	default:
		return "MINIMAL"
	}
}

// FindSimilar finds memories similar to a given memory
func (s *SemanticSearcher) FindSimilar(targetID string, limit int) ([]SearchResult, error) {
	target, err := s.manager.Get(targetID)
	if err != nil {
		return nil, fmt.Errorf("target memory not found: %w", err)
	}

	// Search using target's title and keywords
	query := target.Title + " " + target.Keywords
	results, err := s.Search(query, SearchOptions{
		Limit:      limit + 1,
		MinScore:   0.2,
		MemoryType: target.Type,
	})
	if err != nil {
		return nil, err
	}

	// Filter out the target itself
	var filtered []SearchResult
	for _, r := range results.Results {
		if r.Memory.ID != targetID {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

// SearchByType searches memories by type with scoring
func (s *SemanticSearcher) SearchByType(memType string, limit int) ([]SearchResult, error) {
	memories, err := s.manager.List(memType, limit*2)
	if err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	// Score by importance and recency
	var scored []SearchResult
	for _, mem := range memories {
		score := mem.Importance*0.5 + s.recencyScore(mem.CreatedAt)*0.5
		scored = append(scored, SearchResult{
			Memory:    mem,
			Score:     score,
			Relevance: scoreToRelevance(score),
		})
	}

	// Sort by score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}

	return scored, nil
}
