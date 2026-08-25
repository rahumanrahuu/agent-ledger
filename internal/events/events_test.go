package events

import (
	"os"
	"testing"

	"agent-ledger/internal/storage"
)

func TestEventsManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-ledger-events-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	st := storage.New(tmpDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	manager := NewManager(st)

	sessionID := "test-session"

	// Test Decision
	d, err := manager.CreateDecision(sessionID, "Decision 1", "Use X", "Reason Y", nil, nil)
	if err != nil {
		t.Fatalf("CreateDecision failed: %v", err)
	}
	if d.Title != "Decision 1" {
		t.Errorf("Expected title 'Decision 1', got '%s'", d.Title)
	}

	// Test Discovery
	disc, err := manager.CreateDiscovery(sessionID, "Discovery 1", "Found Z", nil, nil)
	if err != nil {
		t.Fatalf("CreateDiscovery failed: %v", err)
	}
	if disc.Title != "Discovery 1" {
		t.Errorf("Expected title 'Discovery 1', got '%s'", disc.Title)
	}

	// Test Failure
	f, err := manager.CreateFailure(sessionID, "Failure 1", "Try X", "Didn't work", "Lessons")
	if err != nil {
		t.Fatalf("CreateFailure failed: %v", err)
	}
	if f.Title != "Failure 1" {
		t.Errorf("Expected title 'Failure 1', got '%s'", f.Title)
	}

	// Test Constraint
	c, err := manager.CreateConstraint(sessionID, "Constraint 1", "Must do X", "Because Y")
	if err != nil {
		t.Fatalf("CreateConstraint failed: %v", err)
	}
	if c.Title != "Constraint 1" {
		t.Errorf("Expected title 'Constraint 1', got '%s'", c.Title)
	}
	
	// Test Handoff
	h, err := manager.CreateHandoff(sessionID, "State", "Changes", nil, nil, nil, nil, "", "", nil, nil)
	if err != nil {
		t.Fatalf("CreateHandoff failed: %v", err)
	}
	if h.SessionID != sessionID {
		t.Errorf("Expected session ID '%s', got '%s'", sessionID, h.SessionID)
	}
	
	// Format tests are implicitly covered by Parse tests in query_test.go
}
