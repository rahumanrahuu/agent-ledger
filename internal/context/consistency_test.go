package context

import (
	"os"
	"testing"
	"time"

	"agent-ledger/internal/checkpoint"
	"agent-ledger/internal/events"
	"agent-ledger/internal/history"
	"agent-ledger/internal/session"
	"agent-ledger/internal/storage"
)

func TestContextConsistency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-ledger-context-consistency-test-")
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
	historyManager := history.NewManager(sessionManager, checkpointManager, st)
	eventsManager := events.NewManager(st)

	ctxManager := NewManager(historyManager, checkpointManager, st)

	sess, err := sessionManager.Create("test-agent", "test-model", tmpDir, "main", "123456")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create some events
	eventsManager.CreateDecision(sess.ID, "Title 1", "Decision 1", "Rationale 1", nil, nil)
	eventsManager.CreateDiscovery(sess.ID, "Discovery 1", "Finding 1", nil, nil)
	eventsManager.CreateFailure(sess.ID, "Failure 1", "Attempt 1", "Why 1", "Lessons 1")
	eventsManager.CreateConstraint(sess.ID, "Constraint 1", "Constraint 1", "Reason 1")

	// Wait to ensure timestamp diff
	time.Sleep(10 * time.Millisecond)

	// We can't fully compile context without a git repo, so we'll test the internal getters directly
	now := time.Now()
	
	decisions := ctxManager.getDecisions(now, "")
	if len(decisions) != 1 {
		t.Errorf("Expected 1 decision, got %d", len(decisions))
	}
	if len(decisions) > 0 && decisions[0].Title != "Title 1" {
		t.Errorf("Expected decision title 'Title 1', got '%s'", decisions[0].Title)
	}
	if len(decisions) > 0 && decisions[0].Timestamp.IsZero() {
		t.Error("Decision timestamp is zero")
	}

	discoveries := ctxManager.getDiscoveries(now, "")
	if len(discoveries) != 1 {
		t.Errorf("Expected 1 discovery, got %d", len(discoveries))
	}
	if len(discoveries) > 0 && discoveries[0].Timestamp.IsZero() {
		t.Error("Discovery timestamp is zero")
	}

	failures := ctxManager.getFailures(now, "")
	if len(failures) != 1 {
		t.Errorf("Expected 1 failure, got %d", len(failures))
	}
	if len(failures) > 0 && failures[0].Timestamp.IsZero() {
		t.Error("Failure timestamp is zero")
	}

	constraints := ctxManager.getConstraints(now, "")
	if len(constraints) != 1 {
		t.Errorf("Expected 1 constraint, got %d", len(constraints))
	}
	if len(constraints) > 0 && constraints[0].Timestamp.IsZero() {
		t.Error("Constraint timestamp is zero")
	}

	// Test task filtering
	decisionsFiltered := ctxManager.getDecisions(now, "Title 1")
	if len(decisionsFiltered) != 1 {
		t.Errorf("Expected 1 filtered decision, got %d", len(decisionsFiltered))
	}
	
	decisionsFilteredOut := ctxManager.getDecisions(now, "NonExistentTask")
	if len(decisionsFilteredOut) != 0 {
		t.Errorf("Expected 0 filtered decisions, got %d", len(decisionsFilteredOut))
	}
}

func TestGetLatestHandoff(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-ledger-handoff-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	st := storage.New(tmpDir)
	st.Initialize()

	sessionManager := session.NewManager(st)
	ctxManager := NewManager(nil, nil, st)
	eventsManager := events.NewManager(st)

	// Session 1: Started, has handoff, no end time
	sess1, _ := sessionManager.Create("agent1", "model1", tmpDir, "main", "123")
	eventsManager.CreateHandoff(sess1.ID, "State 1", "Changed 1", nil, nil, nil, nil, "", "", nil, nil)
	
	time.Sleep(10 * time.Millisecond) // ensure time passes
	
	// Ensure we get handoff from active session
	handoff1 := ctxManager.getLatestHandoff()
	if handoff1 == "" {
		t.Error("Expected to get handoff from active session")
	}

	// Session 2: Started, has handoff, is stopped
	sess2, _ := sessionManager.Create("agent2", "model2", tmpDir, "main", "456")
	eventsManager.CreateHandoff(sess2.ID, "State 2", "Changed 2", nil, nil, nil, nil, "", "", nil, nil)
	sessionManager.Stop()

	// Should get handoff from Session 2 (newer)
	handoff2 := ctxManager.getLatestHandoff()
	if handoff2 == "" {
		t.Error("Expected to get handoff from stopped session")
	}
	if handoff1 == handoff2 {
		t.Error("Expected newer handoff")
	}
}
