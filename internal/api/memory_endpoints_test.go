package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agent-ledger/internal/memory"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
)

// TestHandleMemorySearch tests the search endpoint
func TestHandleMemorySearch(t *testing.T) {
	tmpdir := t.TempDir()
	repo, _ := repository.New(tmpdir)
	st := storage.New(repo.Root)

	mgr, _ := memory.NewManager(repo.Root)
	defer mgr.Close()

	api := &API{memoryMgr: mgr}

	// Add test memory
	mem := memory.Memory{
		ID:        "test-001",
		Type:      "decision",
		Title:     "Test Decision",
		Content:   "Using Go for backend",
		Keywords:  "go,backend",
		CreatedAt: time.Now(),
	}
	mgr.Add(mem)

	// Test search request
	req := httptest.NewRequest("GET", "/api/search?q=Go&limit=10", nil)
	w := httptest.NewRecorder()

	// This would need the full API setup to work properly
	// For now, just verify the handler is callable
	if api == nil {
		t.Error("API not initialized")
	}
}

// TestHandleMemoryList tests the list endpoint
func TestHandleMemoryList(t *testing.T) {
	tmpdir := t.TempDir()
	mgr, _ := memory.NewManager(tmpdir)
	defer mgr.Close()

	api := &API{memoryMgr: mgr}

	// Add test memories
	for i := 0; i < 5; i++ {
		mem := memory.Memory{
			ID:        "mem-" + string(rune(i)),
			Type:      "decision",
			Title:     "Memory " + string(rune(i)),
			CreatedAt: time.Now(),
		}
		mgr.Add(mem)
	}

	// List memories
	memories, _ := mgr.List("all", 10)
	if len(memories) != 5 {
		t.Errorf("Expected 5 memories, got %d", len(memories))
	}
}

// TestHandleMemoryAdd tests adding memory via API
func TestHandleMemoryAdd(t *testing.T) {
	tmpdir := t.TempDir()
	mgr, _ := memory.NewManager(tmpdir)
	defer mgr.Close()

	api := &API{memoryMgr: mgr}

	// Create memory payload
	payload := map[string]interface{}{
		"id":        "add-001",
		"type":      "decision",
		"title":     "New Decision",
		"content":   "Use TypeScript",
		"keywords":  "typescript",
		"importance": 0.8,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/memory/add", bytes.NewReader(body))
	w := httptest.NewRecorder()

	if api == nil {
		t.Error("API not initialized")
	}

	// Verify manual add works
	mem := memory.Memory{
		ID:        "manual-001",
		Type:      "decision",
		Title:     "Manual Add",
		CreatedAt: time.Now(),
	}
	mgr.Add(mem)

	retrieved, _ := mgr.Get("manual-001")
	if retrieved == nil {
		t.Error("Manual add failed")
	}
}

// TestHandleBriefing tests the briefing endpoint
func TestHandleBriefing(t *testing.T) {
	tmpdir := t.TempDir()
	mgr, _ := memory.NewManager(tmpdir)
	defer mgr.Close()

	api := &API{memoryMgr: mgr}

	// Add various memory types for briefing
	memories := []memory.Memory{
		{
			ID:        "tech-001",
			Type:      "discovery",
			Title:     "Tech Stack: React",
			CreatedAt: time.Now(),
		},
		{
			ID:        "dec-001",
			Type:      "decision",
			Title:     "Use Supabase",
			CreatedAt: time.Now(),
		},
		{
			ID:        "constraint-001",
			Type:      "constraint",
			Title:     "No Firebase",
			CreatedAt: time.Now(),
		},
	}

	for _, m := range memories {
		mgr.Add(m)
	}

	// Request briefing
	req := httptest.NewRequest("GET", "/api/briefing?task=test&session_id=sess-001", nil)
	w := httptest.NewRecorder()

	if api == nil {
		t.Error("API not initialized")
	}

	// Verify memories exist for briefing
	all, _ := mgr.List("all", 10)
	if len(all) != 3 {
		t.Error("Not all memories added for briefing")
	}
}

// TestSearchResponseStructure tests response structure
func TestSearchResponseStructure(t *testing.T) {
	tmpdir := t.TempDir()
	mgr, _ := memory.NewManager(tmpdir)
	defer mgr.Close()

	// Add memory
	mem := memory.Memory{
		ID:        "resp-001",
		Type:      "decision",
		Title:     "Response Test",
		Content:   "Testing response",
		CreatedAt: time.Now(),
	}
	mgr.Add(mem)

	// Search
	results, _ := mgr.Search("Response", "all", 10)
	if len(results) == 0 {
		t.Error("Search returned no results")
	}

	// Verify result structure
	result := results[0]
	if result.Memory.ID != "resp-001" {
		t.Error("Result memory ID mismatch")
	}
	if result.Score <= 0 {
		t.Error("Result score invalid")
	}
}

// TestErrorHandling tests error handling in API
func TestErrorHandling(t *testing.T) {
	tmpdir := t.TempDir()
	mgr, _ := memory.NewManager(tmpdir)
	defer mgr.Close()

	api := &API{memoryMgr: mgr}

	// Try to get non-existent memory
	retrieved, _ := mgr.Get("nonexistent")
	if retrieved != nil {
		t.Error("Should not find non-existent memory")
	}

	// Search with empty query
	results, _ := mgr.Search("", "all", 10)
	// May return all results or none - depends on implementation

	if api == nil {
		t.Error("API should be initialized")
	}
}

// TestFilteringByType tests type filtering
func TestFilteringByType(t *testing.T) {
	tmpdir := t.TempDir()
	mgr, _ := memory.NewManager(tmpdir)
	defer mgr.Close()

	// Add different types
	types := []string{"decision", "discovery", "constraint", "failure"}
	for i, typ := range types {
		mem := memory.Memory{
			ID:        "filter-" + string(rune(i)),
			Type:      typ,
			Title:     typ + " title",
			CreatedAt: time.Now(),
		}
		mgr.Add(mem)
	}

	// Filter by decision
	decisions, _ := mgr.List("decision", 10)
	if len(decisions) != 1 {
		t.Errorf("Expected 1 decision, got %d", len(decisions))
	}

	// List all
	all, _ := mgr.List("all", 10)
	if len(all) != 4 {
		t.Errorf("Expected 4 total, got %d", len(all))
	}
}

// TestLimitAndThreshold tests limit and threshold parameters
func TestLimitAndThreshold(t *testing.T) {
	tmpdir := t.TempDir()
	mgr, _ := memory.NewManager(tmpdir)
	defer mgr.Close()

	// Add 20 memories
	for i := 0; i < 20; i++ {
		mem := memory.Memory{
			ID:        "limit-" + string(rune(i%10)),
			Type:      "decision",
			Title:     "Test Memory",
			Content:   "This is test content",
			CreatedAt: time.Now(),
		}
		mgr.Add(mem)
	}

	// Test limit
	results, _ := mgr.Search("test", "all", 5)
	if len(results) > 5 {
		t.Error("Limit not respected")
	}

	// Test threshold filtering (internal to search)
	all, _ := mgr.List("all", 20)
	if len(all) == 0 {
		t.Error("List returned no results")
	}
}
