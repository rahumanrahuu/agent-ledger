package context

import (
	"os"
	"path/filepath"
	"testing"
	
	"agent-ledger/internal/checkpoint"
	"agent-ledger/internal/events"
	"agent-ledger/internal/history"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/session"
	"agent-ledger/internal/storage"
)

func TestArchitectureAnalysis(t *testing.T) {
	// Create a temporary test repository
	tmpDir, err := os.MkdirTemp("", "agent-ledger-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Initialize a git repository
	repo := repository.Repository{Root: tmpDir}
	
	// Create a test Go project structure
	os.MkdirAll(filepath.Join(tmpDir, "cmd", "ledger"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "cmd", "ledger-mcp"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "internal", "git"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "internal", "session"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "internal", "checkpoint"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "internal", "context"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "mcp"), 0755)
	
	// Create some test files
	os.WriteFile(filepath.Join(tmpDir, "cmd", "ledger", "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "cmd", "ledger-mcp", "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "internal", "git", "git.go"), []byte("package git"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "internal", "session", "session.go"), []byte("package session"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "internal", "checkpoint", "checkpoint.go"), []byte("package checkpoint"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "internal", "context", "context.go"), []byte("package context"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "mcp", "resources.go"), []byte("package mcp"), 0644)
	
	// Initialize agent ledger storage
	st := storage.New(tmpDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	
	// Create managers
	sessionManager := session.NewManager(st)
	checkpointManager := checkpoint.NewManager(st)
	historyManager := history.NewManager(sessionManager, checkpointManager, st)
	contextManager := NewManager(historyManager, checkpointManager, st)
	
	// Compile context
	ctx, err := contextManager.Compile(&repo, "")
	if err != nil {
		t.Fatalf("Failed to compile context: %v", err)
	}
	
	// Verify architecture was analyzed
	if ctx.Architecture == "" {
		t.Error("Architecture should not be empty")
	}
	
	// Verify key components are mentioned
	archStr := ctx.Architecture
	if !contains(archStr, "cmd/") {
		t.Error("Architecture should mention cmd/ directory")
	}
	if !contains(archStr, "internal/") {
		t.Error("Architecture should mention internal/ directory")
	}
	if !contains(archStr, "mcp/") {
		t.Error("Architecture should mention mcp/ directory")
	}
}

func TestRelevanceFiltering(t *testing.T) {
	// Create a temporary test repository
	tmpDir, err := os.MkdirTemp("", "agent-ledger-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Initialize agent ledger storage
	st := storage.New(tmpDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	
	// Create a test session
	sessionManager := session.NewManager(st)
	testSession, err := sessionManager.Create("test-agent", "test-model", tmpDir, "main", "abc123")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	// Create test semantic records
	eventsManager := events.NewManager(st)
	
	// Create an old architectural decision
	oldDecision, err := eventsManager.CreateDecision(testSession.ID, "Use Go for implementation", "Use Go for the core implementation", "Go provides good concurrency support and is widely used", []string{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create old decision: %v", err)
	}
	
	// Create managers
	checkpointManager := checkpoint.NewManager(st)
	historyManager := history.NewManager(sessionManager, checkpointManager, st)
	contextManager := NewManager(historyManager, checkpointManager, st)
	
	// Compile context
	repo := &repository.Repository{Root: tmpDir, Branch: "main", Head: "abc123"}
	ctx, err := contextManager.Compile(repo, "")
	if err != nil {
		t.Fatalf("Failed to compile context: %v", err)
	}
	
	// Verify the architectural decision is included (old but important)
	if len(ctx.ImportantDecisions) == 0 {
		t.Error("Should include at least one decision")
	}
	
	found := false
	for _, dec := range ctx.ImportantDecisions {
		if contains(dec.Title, "Go") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Architectural decision about Go should be included")
	}
	
	_ = oldDecision // Use the variable
}

func TestTaskSpecificContext(t *testing.T) {
	// Create a temporary test repository
	tmpDir, err := os.MkdirTemp("", "agent-ledger-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Initialize agent ledger storage
	st := storage.New(tmpDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	
	// Create managers
	sessionManager := session.NewManager(st)
	checkpointManager := checkpoint.NewManager(st)
	historyManager := history.NewManager(sessionManager, checkpointManager, st)
	contextManager := NewManager(historyManager, checkpointManager, st)
	
	// Compile context with task
	repo := &repository.Repository{Root: tmpDir, Branch: "main", Head: "abc123"}
	ctx, err := contextManager.Compile(repo, "implement MCP resources")
	if err != nil {
		t.Fatalf("Failed to compile context: %v", err)
	}
	
	// Verify task context is present
	if ctx.TaskContext == "" {
		t.Error("Task context should not be empty when task is provided")
	}
	
	if !contains(ctx.TaskContext, "implement MCP resources") {
		t.Error("Task context should mention the provided task")
	}
}

func TestTestingStatus(t *testing.T) {
	// Create a temporary test repository
	tmpDir, err := os.MkdirTemp("", "agent-ledger-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create a test Go project structure with tests
	os.MkdirAll(filepath.Join(tmpDir, "internal", "git"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "internal", "session"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "internal", "checkpoint"), 0755)
	
	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "internal", "git", "git.go"), []byte("package git"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "internal", "git", "git_test.go"), []byte("package git"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "internal", "session", "session.go"), []byte("package session"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "internal", "session", "session_test.go"), []byte("package session"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "internal", "checkpoint", "checkpoint.go"), []byte("package checkpoint"), 0644)
	
	// Initialize agent ledger storage
	st := storage.New(tmpDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	
	// Create managers
	sessionManager := session.NewManager(st)
	checkpointManager := checkpoint.NewManager(st)
	historyManager := history.NewManager(sessionManager, checkpointManager, st)
	contextManager := NewManager(historyManager, checkpointManager, st)
	
	// Compile context
	repo := &repository.Repository{Root: tmpDir, Branch: "main", Head: "abc123"}
	ctx, err := contextManager.Compile(repo, "")
	if err != nil {
		t.Fatalf("Failed to compile context: %v", err)
	}
	
	// Verify testing status is present
	if ctx.TestingStatus == "" {
		t.Error("Testing status should not be empty")
	}
	
	// Verify test files are detected
	if !contains(ctx.TestingStatus, "git") || !contains(ctx.TestingStatus, "session") {
		t.Error("Testing status should mention packages with tests")
	}
}

func TestRecentChanges(t *testing.T) {
	// Create a temporary test repository
	tmpDir, err := os.MkdirTemp("", "agent-ledger-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Initialize agent ledger storage
	st := storage.New(tmpDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	
	// Create managers
	sessionManager := session.NewManager(st)
	checkpointManager := checkpoint.NewManager(st)
	historyManager := history.NewManager(sessionManager, checkpointManager, st)
	contextManager := NewManager(historyManager, checkpointManager, st)
	
	// Compile context with simulated file changes
	repo := &repository.Repository{
		Root:     tmpDir,
		Branch:   "main",
		Head:     "abc123",
		Dirty:    true,
		Staged:   []string{"internal/checkpoint/checkpoint.go"},
		Unstaged: []string{"internal/context/context.go"},
		Untracked: []string{"internal/git/git.go"},
	}
	
	ctx, err := contextManager.Compile(repo, "")
	if err != nil {
		t.Fatalf("Failed to compile context: %v", err)
	}
	
	// Verify recent changes are tracked
	if len(ctx.RecentlyChanged) == 0 {
		t.Error("Should track recent file changes")
	}
	
	// Verify different change types are captured
	foundStaged := false
	foundUnstaged := false
	for _, change := range ctx.RecentlyChanged {
		if change.ChangeType == "modified (staged)" {
			foundStaged = true
		}
		if change.ChangeType == "modified (unstaged)" {
			foundUnstaged = true
		}
	}
	
	if !foundStaged {
		t.Error("Should capture staged changes")
	}
	if !foundUnstaged {
		t.Error("Should capture unstaged changes")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
