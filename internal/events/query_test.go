package events

import (
	"os"
	"testing"
	"time"

	"agent-ledger/internal/storage"
)

func TestParseTimestamp(t *testing.T) {
	// Test standard format
	t1 := parseTimestamp("*Created: 2026-08-25T13:54:41Z*")
	if t1.IsZero() {
		t.Error("Failed to parse standard format")
	}

	// Test without trailing star
	t2 := parseTimestamp("*Created: 2026-08-25T13:54:41Z")
	if t2.IsZero() {
		t.Error("Failed to parse format without trailing star")
	}

	// Test with spaces
	t3 := parseTimestamp("*Created:   2026-08-25T13:54:41Z  *")
	if t3.IsZero() {
		t.Error("Failed to parse format with spaces")
	}
}

func TestQueryLayer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-ledger-query-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	st := storage.New(tmpDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	manager := NewManager(st)

	// Create some records
	sessionID := "test-session"
	_, err = manager.CreateDecision(sessionID, "Decision 1", "Use Go", "Because it's fast", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create decision: %v", err)
	}
	time.Sleep(10 * time.Millisecond) // Ensure timestamps differ slightly
	_, err = manager.CreateDecision(sessionID, "Decision 2", "Use JSON", "Simple", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create decision: %v", err)
	}

	_, err = manager.CreateDiscovery(sessionID, "Discovery 1", "Found a bug", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create discovery: %v", err)
	}

	_, err = manager.CreateFailure(sessionID, "Failure 1", "Try X", "Didn't work", "Don't do X")
	if err != nil {
		t.Fatalf("Failed to create failure: %v", err)
	}

	_, err = manager.CreateConstraint(sessionID, "Constraint 1", "Must use Go 1.21+", "Features")
	if err != nil {
		t.Fatalf("Failed to create constraint: %v", err)
	}

	// Test ListDecisions
	decisions, err := ListDecisions(st)
	if err != nil {
		t.Fatalf("ListDecisions failed: %v", err)
	}
	if len(decisions) != 2 {
		t.Errorf("Expected 2 decisions, got %d", len(decisions))
	}
	if decisions[0].Title != "Decision 2" { // Should be sorted newest first
		t.Errorf("Expected newest decision first (Decision 2), got %s", decisions[0].Title)
	}
	if CountDecisions(st) != 2 {
		t.Errorf("CountDecisions should return 2")
	}

	// Test ListDiscoveries
	discoveries, err := ListDiscoveries(st)
	if err != nil {
		t.Fatalf("ListDiscoveries failed: %v", err)
	}
	if len(discoveries) != 1 {
		t.Errorf("Expected 1 discovery, got %d", len(discoveries))
	}
	if CountDiscoveries(st) != 1 {
		t.Errorf("CountDiscoveries should return 1")
	}

	// Test ListFailures
	failures, err := ListFailures(st)
	if err != nil {
		t.Fatalf("ListFailures failed: %v", err)
	}
	if len(failures) != 1 {
		t.Errorf("Expected 1 failure, got %d", len(failures))
	}
	if CountFailures(st) != 1 {
		t.Errorf("CountFailures should return 1")
	}

	// Test ListConstraints
	constraints, err := ListConstraints(st)
	if err != nil {
		t.Fatalf("ListConstraints failed: %v", err)
	}
	if len(constraints) != 1 {
		t.Errorf("Expected 1 constraint, got %d", len(constraints))
	}
	if CountConstraints(st) != 1 {
		t.Errorf("CountConstraints should return 1")
	}
}

func TestParseInvalidMarkdown(t *testing.T) {
	invalidContent := `This is not a valid
record format at all.`
	
	if d := ParseDecision(invalidContent); d != nil {
		t.Error("ParseDecision should return nil for invalid content")
	}
	if d := ParseDiscovery(invalidContent); d != nil {
		t.Error("ParseDiscovery should return nil for invalid content")
	}
	if f := ParseFailure(invalidContent); f != nil {
		t.Error("ParseFailure should return nil for invalid content")
	}
	if c := ParseConstraint(invalidContent); c != nil {
		t.Error("ParseConstraint should return nil for invalid content")
	}
}
