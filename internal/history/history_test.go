package history

import (
	"os"
	"testing"

	"agent-ledger/internal/checkpoint"
	"agent-ledger/internal/events"
	"agent-ledger/internal/session"
	"agent-ledger/internal/storage"
)

func TestHistoryManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-ledger-history-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	st := storage.New(tmpDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	sessionManager := session.NewManager(st)
	checkpointManager := checkpoint.NewManager(st)
	manager := NewManager(sessionManager, checkpointManager, st)

	sess1, err := sessionManager.Create("agent1", "model1", tmpDir, "main", "head")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	// Add some events to session 1
	eventsManager := events.NewManager(st)
	eventsManager.CreateDecision(sess1.ID, "Title 1", "Decision 1", "Rationale 1", nil, nil)

	// Create session 2
	_, err = sessionManager.Create("agent2", "model2", tmpDir, "main", "head2")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	// Test GetAllSessions
	all, err := manager.GetAllSessions("", "")
	if err != nil {
		t.Fatalf("GetAllSessions failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(all))
	}
	
	// Test GetAllSessions with filter
	filtered, err := manager.GetAllSessions("agent1", "")
	if err != nil {
		t.Fatalf("GetAllSessions with filter failed: %v", err)
	}
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered session, got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].ID != sess1.ID {
		t.Errorf("Expected session 1, got %s", filtered[0].ID)
	}
	
	// Test GetSessionHistory
	history, err := manager.GetSessionHistory(sess1.ID)
	if err != nil {
		t.Fatalf("GetSessionHistory failed: %v", err)
	}
	if history.Session.ID != sess1.ID {
		t.Errorf("Expected session ID %s, got %s", sess1.ID, history.Session.ID)
	}
	if len(history.Decisions) != 1 {
		t.Errorf("Expected 1 decision, got %d", len(history.Decisions))
	}
}
