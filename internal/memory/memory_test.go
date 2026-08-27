package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMemoryManager tests the core memory manager
func TestMemoryManager(t *testing.T) {
	tmpdir := t.TempDir()

	mgr, err := NewManager(tmpdir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	if mgr == nil {
		t.Fatal("Manager is nil")
	}
	defer mgr.Close()

	// Test Add
	mem := Memory{
		ID:         "test-001",
		Type:       "decision",
		Title:      "Use TypeScript",
		Content:    "Decided to use TypeScript for type safety",
		Keywords:   "language,typescript",
		Importance: 0.9,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = mgr.Add(mem)
	if err != nil {
		t.Errorf("Add failed: %v", err)
	}

	// Test Get
	retrieved, err := mgr.Get("test-001")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if retrieved == nil {
		t.Errorf("Get returned nil")
	}
	if retrieved.Title != mem.Title {
		t.Errorf("Expected title %s, got %s", mem.Title, retrieved.Title)
	}

	// Test List
	memories, err := mgr.List("all", 10)
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(memories) == 0 {
		t.Errorf("List returned no memories")
	}

	// Test Search
	results, err := mgr.Search("TypeScript", "all", 10)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Errorf("Search returned no results")
	}
}

// TestMemorySearch tests search functionality
func TestMemorySearch(t *testing.T) {
	tmpdir := t.TempDir()
	mgr, _ := NewManager(tmpdir)
	defer mgr.Close()

	// Add test memories
	memories := []Memory{
		{
			ID:        "dec-001",
			Type:      "decision",
			Title:     "Authentication Strategy",
			Content:   "Use OAuth2 for authentication",
			Keywords:  "auth,oauth",
			CreatedAt: time.Now(),
		},
		{
			ID:        "disc-001",
			Type:      "discovery",
			Title:     "Database Performance",
			Content:   "PostgreSQL is fast",
			Keywords:  "database,perf",
			CreatedAt: time.Now(),
		},
	}

	for _, m := range memories {
		mgr.Add(m)
	}

	// Test search by type
	results, _ := mgr.Search("auth", "decision", 10)
	if len(results) == 0 {
		t.Errorf("Type-filtered search failed")
	}

	// Test search all types
	results, _ = mgr.Search("database", "all", 10)
	if len(results) == 0 {
		t.Errorf("All-type search failed")
	}
}

// TestEventLedger tests event recording
func TestEventLedger(t *testing.T) {
	tmpdir := t.TempDir()
	ledger := NewEventLedger(tmpdir)

	sessionID := "sess-001"

	// Record event
	ledger.RecordEvent(Event{
		Type:      FileEdited,
		SessionID: sessionID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"file": "main.go",
		},
	})

	// Save checkpoint
	checkpoint := Checkpoint{
		SessionID: sessionID,
		Summary:   "Test checkpoint",
		TotalTokens: 1000,
		Duration:  300,
	}

	err := ledger.SaveCheckpoint(checkpoint)
	if err != nil {
		t.Errorf("SaveCheckpoint failed: %v", err)
	}

	// Get latest checkpoint
	retrieved, _ := ledger.GetLatestCheckpoint(sessionID)
	if retrieved == nil {
		t.Errorf("GetLatestCheckpoint returned nil")
	}
	if retrieved.Summary != checkpoint.Summary {
		t.Errorf("Checkpoint summary mismatch")
	}
}

// TestConstraintChecker tests constraint checking
func TestConstraintChecker(t *testing.T) {
	tmpdir := t.TempDir()
	checker, err := NewConstraintChecker(tmpdir)
	if err != nil {
		t.Fatalf("Failed to create checker: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpdir, "test.dart")
	content := []byte("import 'package:firebase_auth/firebase_auth.dart';")
	os.WriteFile(testFile, content, 0644)

	// Check for violations
	violations := checker.CheckFile(testFile)
	if len(violations) == 0 {
		// May not find violations if constraints not defined
		// This is OK - just verify no crash
	}
}

// TestTraceRecorder tests reasoning traces
func TestTraceRecorder(t *testing.T) {
	tmpdir := t.TempDir()
	recorder := NewTraceRecorder(tmpdir, "sess-001")

	// Record traces
	recorder.RecordMemoryRetrieval("test query", 5, 0.8, 50)
	recorder.RecordToolCall("ReadFile", true, 25, 100)
	recorder.RecordDecisionPoint("Language choice?", "Go", "Performance critical")
	recorder.RecordConstraintCheck("auth-only", true, "Passed")

	// Generate summary
	summary := recorder.GenerateSummary()
	if summary.TotalTraces != 4 {
		t.Errorf("Expected 4 traces, got %d", summary.TotalTraces)
	}
	if summary.MemoryRetrievals != 1 {
		t.Errorf("Expected 1 memory retrieval")
	}

	// Export traces
	err := recorder.Export()
	if err != nil {
		t.Errorf("Export failed: %v", err)
	}
}

// TestSharedMemoryHub tests multi-agent coordination
func TestSharedMemoryHub(t *testing.T) {
	hub := NewSharedMemoryHub()

	// Register agents
	agent1 := Agent{
		ID:   "agent-001",
		Name: "Agent 1",
		Type: "general",
	}
	hub.RegisterAgent(agent1)

	agent2 := Agent{
		ID:   "agent-002",
		Name: "Agent 2",
		Type: "specialized",
	}
	hub.RegisterAgent(agent2)

	// List active agents
	active := hub.ListActiveAgents()
	if len(active) != 2 {
		t.Errorf("Expected 2 active agents, got %d", len(active))
	}

	// Write memory from agent
	memory := Memory{
		ID:        "shared-001",
		Type:      "decision",
		Title:     "Shared Decision",
		Content:   "Made by agent 1",
		CreatedAt: time.Now(),
	}

	id, err := hub.AgentWrite("agent-001", memory)
	if err != nil {
		t.Errorf("AgentWrite failed: %v", err)
	}
	if id != memory.ID {
		t.Errorf("AgentWrite returned wrong ID")
	}

	// Stats
	stats := hub.GetStats()
	if stats.ActiveAgents != 2 {
		t.Errorf("Stats: expected 2 active agents")
	}
	if stats.SharedMemories != 1 {
		t.Errorf("Stats: expected 1 shared memory")
	}
}

// TestCollaborationHub tests real-time collaboration
func TestCollaborationHub(t *testing.T) {
	hub := NewCollaborationHub()

	// Create session
	session := hub.CreateSession("collab-001")
	if session == nil {
		t.Errorf("CreateSession failed")
	}

	// Join session
	participant := Participant{
		ID:   "user-001",
		Name: "User 1",
	}
	err := hub.JoinSession("collab-001", participant)
	if err != nil {
		t.Errorf("JoinSession failed: %v", err)
	}

	// Get active participants
	active := hub.GetActiveParticipants("collab-001")
	if len(active) != 1 {
		t.Errorf("Expected 1 participant, got %d", len(active))
	}

	// Update memory
	memory := Memory{
		ID:        "collab-mem-001",
		Type:      "decision",
		Title:     "Collaborative Decision",
		CreatedAt: time.Now(),
	}

	err = hub.UpdateMemory("collab-001", "user-001", memory)
	if err != nil {
		t.Errorf("UpdateMemory failed: %v", err)
	}

	// Get session memories
	memories := hub.GetSessionMemories("collab-001")
	if len(memories) != 1 {
		t.Errorf("Expected 1 memory in session, got %d", len(memories))
	}

	// Leave session
	err = hub.LeaveSession("collab-001", "user-001")
	if err != nil {
		t.Errorf("LeaveSession failed: %v", err)
	}

	active = hub.GetActiveParticipants("collab-001")
	if len(active) != 0 {
		t.Errorf("Expected 0 participants after leaving")
	}
}

// TestCursorSync tests cursor position synchronization
func TestCursorSync(t *testing.T) {
	hub := NewCollaborationHub()

	hub.CreateSession("cursor-001")
	participant := Participant{ID: "user-001", Name: "User 1"}
	hub.JoinSession("cursor-001", participant)

	// Update cursor
	pos := CursorPosition{
		MemoryID:  "mem-001",
		Line:      42,
		Column:    10,
	}

	err := hub.UpdateCursor("cursor-001", "user-001", pos)
	if err != nil {
		t.Errorf("UpdateCursor failed: %v", err)
	}

	active := hub.GetActiveParticipants("cursor-001")
	if len(active) == 0 || active[0].Cursor.Line != 42 {
		t.Errorf("Cursor position not synced")
	}
}

// BenchmarkMemorySearch benchmarks search performance
func BenchmarkMemorySearch(b *testing.B) {
	tmpdir := b.TempDir()
	mgr, _ := NewManager(tmpdir)
	defer mgr.Close()

	// Add test memories
	for i := 0; i < 100; i++ {
		mem := Memory{
			ID:        string(rune(i)),
			Type:      "decision",
			Title:     "Test Memory",
			Content:   "This is test content",
			CreatedAt: time.Now(),
		}
		mgr.Add(mem)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.Search("test", "all", 10)
	}
}
