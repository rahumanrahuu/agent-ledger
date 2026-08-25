package session

import (
	"testing"
	"time"
	
	"agent-ledger/internal/storage"
)

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestCreateSession(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	
	session, err := manager.Create("test-agent", "gpt-4", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}
	if session.Agent != "test-agent" {
		t.Errorf("Expected agent test-agent, got %s", session.Agent)
	}
	if session.Model != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %s", session.Model)
	}
	if session.RepoRoot != "/test/repo" {
		t.Errorf("Expected repo root /test/repo, got %s", session.RepoRoot)
	}
	if session.Branch != "main" {
		t.Errorf("Expected branch main, got %s", session.Branch)
	}
	if session.Head != "abc123" {
		t.Errorf("Expected head abc123, got %s", session.Head)
	}
	if time.Since(session.StartTime) > time.Second {
		t.Error("Start time should be recent")
	}
}

func TestCreateSessionWithArbitraryAgent(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	
	// Test with various arbitrary agent names
	testAgents := []string{
		"custom-agent-v1",
		"internal-tool-2025",
		"future-agent-x",
		"my-company-internal",
		"test-beta-2.0",
	}
	
	for _, agent := range testAgents {
		session, err := manager.Create(agent, "", "/test/repo", "main", "abc123")
		if err != nil {
			t.Fatalf("Create failed for agent %s: %v", agent, err)
		}
		if session.Agent != agent {
			t.Errorf("Expected agent %s, got %s", agent, session.Agent)
		}
	}
}

func TestCreateSessionWithNoAgent(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	
	session, err := manager.Create("", "", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	if session.Agent != "" {
		t.Error("Agent should be empty when not specified")
	}
}

func TestGetSession(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	
	// Create a session
	createdSession, err := manager.Create("test-agent", "gpt-4", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	// Get the session
	retrievedSession, err := manager.Get(createdSession.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	
	if retrievedSession.ID != createdSession.ID {
		t.Errorf("Expected ID %s, got %s", createdSession.ID, retrievedSession.ID)
	}
	if retrievedSession.Agent != createdSession.Agent {
		t.Errorf("Expected agent %s, got %s", createdSession.Agent, retrievedSession.Agent)
	}
}

func TestGetCurrentSession(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	
	// No current session initially
	_, err := manager.GetCurrent()
	if err == nil {
		t.Error("GetCurrent should return error when no current session")
	}
	
	// Create a session
	session, err := manager.Create("test-agent", "gpt-4", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	// Get current session
	currentSession, err := manager.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}
	
	if currentSession.ID != session.ID {
		t.Errorf("Expected current session ID %s, got %s", session.ID, currentSession.ID)
	}
}

func TestSetCurrentSession(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	
	// Create two sessions
	_, err := manager.Create("agent-1", "gpt-4", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	session2, err := manager.Create("agent-2", "gpt-4", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	// Set current to session2
	err = manager.SetCurrent(session2.ID)
	if err != nil {
		t.Fatalf("SetCurrent failed: %v", err)
	}
	
	// Verify current session
	currentSession, err := manager.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}
	
	if currentSession.ID != session2.ID {
		t.Errorf("Expected current session ID %s, got %s", session2.ID, currentSession.ID)
	}
}

func TestClearCurrentSession(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	
	// Create and set current session
	_, err := manager.Create("test-agent", "gpt-4", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	// Clear current session
	err = manager.ClearCurrent()
	if err != nil {
		t.Fatalf("ClearCurrent failed: %v", err)
	}
	
	// Verify no current session
	_, err = manager.GetCurrent()
	if err == nil {
		t.Error("GetCurrent should return error after ClearCurrent")
	}
}

func TestStopSession(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	
	// Create a session
	session, err := manager.Create("test-agent", "gpt-4", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	// Stop the session
	err = manager.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	
	// Verify session was stopped
	retrievedSession, err := manager.Get(session.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	
	if retrievedSession.EndTime == nil {
		t.Error("EndTime should be set after stop")
	}
	
	// Verify current session was cleared
	_, err = manager.GetCurrent()
	if err == nil {
		t.Error("GetCurrent should return error after stop")
	}
}

func TestListAllSessions(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	
	// Create multiple sessions
	_, err := manager.Create("agent-1", "gpt-4", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	_, err = manager.Create("agent-2", "gpt-4", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	_, err = manager.Create("agent-3", "gpt-4", "/test/repo", "main", "abc123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	// List all sessions
	sessions, err := manager.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	
	if len(sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(sessions))
	}
}
