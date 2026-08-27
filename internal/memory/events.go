package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// EventType defines the type of event
type EventType string

const (
	SessionStarted       EventType = "session_started"
	FileRead             EventType = "file_read"
	FileEdited           EventType = "file_edited"
	DecisionRecorded     EventType = "decision_recorded"
	DiscoveryRecorded    EventType = "discovery_recorded"
	ToolCall             EventType = "tool_call"
	ErrorOccurred        EventType = "error_occurred"
	ConstraintViolation  EventType = "constraint_violated"
	SessionCheckpoint    EventType = "session_checkpoint"
)

// Event represents a single event in the ledger
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	SessionID string                 `json:"session_id"`
	Data      map[string]interface{} `json:"data"`
}

// EventLedger manages event recording
type EventLedger struct {
	root string
}

// NewEventLedger creates a new event ledger
func NewEventLedger(root string) *EventLedger {
	return &EventLedger{root: root}
}

// RecordEvent records a new event
func (el *EventLedger) RecordEvent(event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Ensure session directory exists
	sessionDir := filepath.Join(el.root, ".agent", "events", event.SessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}

	// Append to session events file
	eventsFile := filepath.Join(sessionDir, "events.jsonl")
	f, err := os.OpenFile(eventsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	return encoder.Encode(event)
}

// GetSessionEvents retrieves all events for a session
func (el *EventLedger) GetSessionEvents(sessionID string) ([]Event, error) {
	eventsFile := filepath.Join(el.root, ".agent", "events", sessionID, "events.jsonl")

	f, err := os.Open(eventsFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []Event
	decoder := json.NewDecoder(f)
	for decoder.Decode(&events) != nil {
		// Intentional - read until EOF
	}

	return events, nil
}

// CreateCheckpoint creates a session checkpoint with summary
type Checkpoint struct {
	SessionID       string            `json:"session_id"`
	Timestamp       time.Time         `json:"timestamp"`
	Summary         string            `json:"summary"`
	FilesModified   []string          `json:"files_modified"`
	DecisionsMade   []string          `json:"decisions_made"`
	DiscoveriesMade []string          `json:"discoveries_made"`
	NextSteps       []string          `json:"next_steps"`
	TotalTokens     int               `json:"total_tokens"`
	Duration        int               `json:"duration_seconds"`
	Metadata        map[string]string `json:"metadata"`
}

// SaveCheckpoint saves a checkpoint
func (el *EventLedger) SaveCheckpoint(checkpoint Checkpoint) error {
	if checkpoint.Timestamp.IsZero() {
		checkpoint.Timestamp = time.Now()
	}

	// Create checkpoint directory
	checkpointDir := filepath.Join(el.root, ".agent", "events", checkpoint.SessionID, "checkpoints")
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		return err
	}

	// Save checkpoint with timestamp as filename
	filename := checkpoint.Timestamp.Format("20060102-150405") + ".json"
	filepath := filepath.Join(checkpointDir, filename)

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// GetLatestCheckpoint retrieves the most recent checkpoint
func (el *EventLedger) GetLatestCheckpoint(sessionID string) (*Checkpoint, error) {
	checkpointDir := filepath.Join(el.root, ".agent", "events", sessionID, "checkpoints")

	entries, err := os.ReadDir(checkpointDir)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, nil // No checkpoints yet
	}

	// Get the last file (most recent)
	lastFile := entries[len(entries)-1]
	filepath := filepath.Join(checkpointDir, lastFile.Name())

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, err
	}

	return &checkpoint, nil
}

// GenerateSessionSummary creates a markdown summary from events
func (el *EventLedger) GenerateSessionSummary(sessionID string) (string, error) {
	checkpoint, err := el.GetLatestCheckpoint(sessionID)
	if err != nil || checkpoint == nil {
		return "", err
	}

	summary := "# Session Summary\n\n"
	summary += "**Session ID:** " + sessionID + "\n"
	summary += "**Timestamp:** " + checkpoint.Timestamp.Format("2006-01-02 15:04:05") + "\n"
	summary += "**Duration:** " + formatDuration(checkpoint.Duration) + "\n"
	summary += "**Tokens Used:** " + string(rune(checkpoint.TotalTokens)) + "\n\n"

	summary += "## Overview\n" + checkpoint.Summary + "\n\n"

	if len(checkpoint.DecisionsMade) > 0 {
		summary += "## Decisions\n"
		for _, d := range checkpoint.DecisionsMade {
			summary += "- " + d + "\n"
		}
		summary += "\n"
	}

	if len(checkpoint.DiscoveriesMade) > 0 {
		summary += "## Discoveries\n"
		for _, d := range checkpoint.DiscoveriesMade {
			summary += "- " + d + "\n"
		}
		summary += "\n"
	}

	if len(checkpoint.FilesModified) > 0 {
		summary += "## Files Modified\n"
		for _, f := range checkpoint.FilesModified {
			summary += "- " + f + "\n"
		}
		summary += "\n"
	}

	if len(checkpoint.NextSteps) > 0 {
		summary += "## Next Steps\n"
		for _, s := range checkpoint.NextSteps {
			summary += "- " + s + "\n"
		}
		summary += "\n"
	}

	return summary, nil
}

func formatDuration(seconds int) string {
	mins := seconds / 60
	secs := seconds % 60
	return string(rune(mins)) + "m " + string(rune(secs)) + "s"
}
