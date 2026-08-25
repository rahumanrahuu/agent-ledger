package events

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/google/uuid"
	"agent-ledger/internal/storage"
)

// BaseEvent represents a base event
type BaseEvent struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
	Agent     string    `json:"agent,omitempty"`
	Model     string    `json:"model,omitempty"`
}

// Decision represents a decision record
type Decision struct {
	BaseEvent
	Title         string   `json:"title"`
	Decision      string   `json:"decision"`
	Rationale      string   `json:"rationale"`
	Alternatives  []string `json:"alternatives,omitempty"`
	Evidence      string   `json:"evidence,omitempty"`
	RelatedFiles  []string `json:"related_files,omitempty"`
	RelatedPackages []string `json:"related_packages,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// Discovery represents a discovery record
type Discovery struct {
	BaseEvent
	Title          string   `json:"title"`
	Finding        string   `json:"finding"`
	Evidence       string   `json:"evidence,omitempty"`
	AffectedAreas  []string `json:"affected_areas,omitempty"`
	RelatedFiles   []string `json:"related_files,omitempty"`
	RelatedPackages []string `json:"related_packages,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// Failure represents a failure record
type Failure struct {
	BaseEvent
	Title            string   `json:"title"`
	AttemptedApproach string `json:"attempted_approach"`
	WhyItFailed      string `json:"why_it_failed"`
	Lessons          string `json:"lessons"`
	RelatedFiles     []string `json:"related_files,omitempty"`
	RelatedPackages  []string `json:"related_packages,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// Constraint represents a constraint record
type Constraint struct {
	BaseEvent
	Title     string `json:"title"`
	Constraint string `json:"constraint"`
	Reason    string `json:"reason"`
}

// Handoff represents a handoff record
type Handoff struct {
	BaseEvent
	CurrentState       string   `json:"current_state"`
	WhatChanged        string   `json:"what_changed"`
	ImportantDecisions []string `json:"important_decisions,omitempty"`
	Discoveries        []string `json:"discoveries,omitempty"`
	Failures           []string `json:"failures,omitempty"`
	Constraints        []string `json:"constraints,omitempty"`
	UnresolvedWork     string   `json:"unresolved_work,omitempty"`
	RecommendedNextSteps string `json:"recommended_next_steps,omitempty"`
	RelevantFiles      []string `json:"relevant_files,omitempty"`
	RelevantCheckpoints []string `json:"relevant_checkpoints,omitempty"`
}

// Manager manages events
type Manager struct {
	storage *storage.Storage
}

// NewManager creates a new events manager
func NewManager(st *storage.Storage) *Manager {
	return &Manager{
		storage: st,
	}
}

// slugify creates a URL-friendly slug from a string
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	// Remove special characters
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// CreateDecision creates a decision record
func (m *Manager) CreateDecision(sessionID, title, decision, rationale string, alternatives, evidence []string) (*Decision, error) {
	id := uuid.New().String()
	slug := slugify(title)
	
	// Get current session if available for agent/model info
	var agent, model string
	sessionPath := fmt.Sprintf("sessions/%s/metadata.json", sessionID)
	var sessionData struct {
		Agent string `json:"agent,omitempty"`
		Model string `json:"model,omitempty"`
	}
	if m.storage.FileExists(sessionPath) {
		m.storage.ReadJSON(sessionPath, &sessionData)
		agent = sessionData.Agent
		model = sessionData.Model
	}
	
	event := &Decision{
		BaseEvent: BaseEvent{
			ID:        id,
			SessionID: sessionID,
			Timestamp: time.Now().UTC(),
			Agent:     agent,
			Model:     model,
		},
		Title:       title,
		Decision:    decision,
		Rationale:   rationale,
		Alternatives: alternatives,
		Evidence:    strings.Join(evidence, "\n"),
	}
	
	// Create Markdown content
	content := m.formatDecision(event)
	filename := fmt.Sprintf("%s-%s.md", id[:8], slug)
	path := fmt.Sprintf("decisions/%s", filename)
	
	if err := m.storage.WriteMarkdown(path, content); err != nil {
		return nil, fmt.Errorf("failed to write decision: %w", err)
	}
	
	return event, nil
}

// formatDecision formats a decision as Markdown
func (m *Manager) formatDecision(d *Decision) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("# %s\n\n", d.Title))
	sb.WriteString(fmt.Sprintf("**Decision:** %s\n\n", d.Decision))
	sb.WriteString(fmt.Sprintf("**Rationale:** %s\n\n", d.Rationale))
	
	if len(d.Alternatives) > 0 {
		sb.WriteString("**Alternatives:**\n")
		for _, alt := range d.Alternatives {
			sb.WriteString(fmt.Sprintf("- %s\n", alt))
		}
		sb.WriteString("\n")
	}
	
	if d.Evidence != "" {
		sb.WriteString(fmt.Sprintf("**Evidence:**\n%s\n", d.Evidence))
	}
	
	sb.WriteString(fmt.Sprintf("\n*Session: %s*\n", d.SessionID))
	sb.WriteString(fmt.Sprintf("*Created: %s*\n", d.Timestamp.Format(time.RFC3339)))
	
	return sb.String()
}

// CreateDiscovery creates a discovery record
func (m *Manager) CreateDiscovery(sessionID, title, finding string, evidence []string, affectedAreas []string) (*Discovery, error) {
	id := uuid.New().String()
	slug := slugify(title)
	
	// Get current session if available for agent/model info
	var agent, model string
	sessionPath := fmt.Sprintf("sessions/%s/metadata.json", sessionID)
	var sessionData struct {
		Agent string `json:"agent,omitempty"`
		Model string `json:"model,omitempty"`
	}
	if m.storage.FileExists(sessionPath) {
		m.storage.ReadJSON(sessionPath, &sessionData)
		agent = sessionData.Agent
		model = sessionData.Model
	}
	
	event := &Discovery{
		BaseEvent: BaseEvent{
			ID:        id,
			SessionID: sessionID,
			Timestamp: time.Now().UTC(),
			Agent:     agent,
			Model:     model,
		},
		Title:         title,
		Finding:       finding,
		Evidence:      strings.Join(evidence, "\n"),
		AffectedAreas: affectedAreas,
	}
	
	// Create Markdown content
	content := m.formatDiscovery(event)
	filename := fmt.Sprintf("%s-%s.md", id[:8], slug)
	path := fmt.Sprintf("discoveries/%s", filename)
	
	if err := m.storage.WriteMarkdown(path, content); err != nil {
		return nil, fmt.Errorf("failed to write discovery: %w", err)
	}
	
	return event, nil
}

// formatDiscovery formats a discovery as Markdown
func (m *Manager) formatDiscovery(d *Discovery) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("# %s\n\n", d.Title))
	sb.WriteString(fmt.Sprintf("**Finding:** %s\n\n", d.Finding))
	
	if len(d.Evidence) > 0 {
		sb.WriteString("**Evidence:**\n")
		for _, ev := range strings.Split(d.Evidence, "\n") {
			if ev != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", ev))
			}
		}
		sb.WriteString("\n")
	}
	
	if len(d.AffectedAreas) > 0 {
		sb.WriteString("**Affected Areas:**\n")
		for _, area := range d.AffectedAreas {
			sb.WriteString(fmt.Sprintf("- %s\n", area))
		}
		sb.WriteString("\n")
	}
	
	sb.WriteString(fmt.Sprintf("*Session: %s*\n", d.SessionID))
	sb.WriteString(fmt.Sprintf("*Created: %s*\n", d.Timestamp.Format(time.RFC3339)))
	
	return sb.String()
}

// CreateFailure creates a failure record
func (m *Manager) CreateFailure(sessionID, title, attemptedApproach, whyItFailed, lessons string) (*Failure, error) {
	id := uuid.New().String()
	slug := slugify(title)
	
	// Get current session if available for agent/model info
	var agent, model string
	sessionPath := fmt.Sprintf("sessions/%s/metadata.json", sessionID)
	var sessionData struct {
		Agent string `json:"agent,omitempty"`
		Model string `json:"model,omitempty"`
	}
	if m.storage.FileExists(sessionPath) {
		m.storage.ReadJSON(sessionPath, &sessionData)
		agent = sessionData.Agent
		model = sessionData.Model
	}
	
	event := &Failure{
		BaseEvent: BaseEvent{
			ID:        id,
			SessionID: sessionID,
			Timestamp: time.Now().UTC(),
			Agent:     agent,
			Model:     model,
		},
		Title:            title,
		AttemptedApproach: attemptedApproach,
		WhyItFailed:     whyItFailed,
		Lessons:         lessons,
	}
	
	// Create Markdown content
	content := m.formatFailure(event)
	filename := fmt.Sprintf("%s-%s.md", id[:8], slug)
	path := fmt.Sprintf("failures/%s", filename)
	
	if err := m.storage.WriteMarkdown(path, content); err != nil {
		return nil, fmt.Errorf("failed to write failure: %w", err)
	}
	
	return event, nil
}

// formatFailure formats a failure as Markdown
func (m *Manager) formatFailure(f *Failure) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("# %s\n\n", f.Title))
	sb.WriteString(fmt.Sprintf("**Attempted Approach:** %s\n\n", f.AttemptedApproach))
	sb.WriteString(fmt.Sprintf("**Why It Failed:** %s\n\n", f.WhyItFailed))
	sb.WriteString(fmt.Sprintf("**Lessons:** %s\n\n", f.Lessons))
	
	sb.WriteString(fmt.Sprintf("*Session: %s*\n", f.SessionID))
	if f.Agent != "" {
		sb.WriteString(fmt.Sprintf("*Agent: %s*\n", f.Agent))
	}
	if f.Model != "" {
		sb.WriteString(fmt.Sprintf("*Model: %s*\n", f.Model))
	}
	sb.WriteString(fmt.Sprintf("*Created: %s*\n", f.Timestamp.Format(time.RFC3339)))
	
	return sb.String()
}

// CreateConstraint creates a constraint record
func (m *Manager) CreateConstraint(sessionID, title, constraint, reason string) (*Constraint, error) {
	id := uuid.New().String()
	slug := slugify(title)
	
	// Get current session if available for agent/model info
	var agent, model string
	sessionPath := fmt.Sprintf("sessions/%s/metadata.json", sessionID)
	var sessionData struct {
		Agent string `json:"agent,omitempty"`
		Model string `json:"model,omitempty"`
	}
	if m.storage.FileExists(sessionPath) {
		m.storage.ReadJSON(sessionPath, &sessionData)
		agent = sessionData.Agent
		model = sessionData.Model
	}
	
	event := &Constraint{
		BaseEvent: BaseEvent{
			ID:        id,
			SessionID: sessionID,
			Timestamp: time.Now().UTC(),
			Agent:     agent,
			Model:     model,
		},
		Title:      title,
		Constraint: constraint,
		Reason:     reason,
	}
	
	// Create Markdown content
	content := m.formatConstraint(event)
	filename := fmt.Sprintf("%s-%s.md", id[:8], slug)
	path := fmt.Sprintf("constraints/%s", filename)
	
	if err := m.storage.WriteMarkdown(path, content); err != nil {
		return nil, fmt.Errorf("failed to write constraint: %w", err)
	}
	
	return event, nil
}

// formatConstraint formats a constraint as Markdown
func (m *Manager) formatConstraint(c *Constraint) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("# %s\n\n", c.Title))
	sb.WriteString(fmt.Sprintf("**Constraint:** %s\n\n", c.Constraint))
	sb.WriteString(fmt.Sprintf("**Reason:** %s\n\n", c.Reason))
	
	sb.WriteString(fmt.Sprintf("*Session: %s*\n", c.SessionID))
	sb.WriteString(fmt.Sprintf("*Created: %s*\n", c.Timestamp.Format(time.RFC3339)))
	
	return sb.String()
}

// CreateHandoff creates a handoff record
func (m *Manager) CreateHandoff(sessionID, currentState, whatChanged string, importantDecisions, discoveries, failures, constraints []string, unresolvedWork, recommendedNextSteps string, relevantFiles, relevantCheckpoints []string) (*Handoff, error) {
	id := uuid.New().String()
	
	event := &Handoff{
		BaseEvent: BaseEvent{
			ID:        id,
			SessionID: sessionID,
			Timestamp: time.Now().UTC(),
		},
		CurrentState:         currentState,
		WhatChanged:         whatChanged,
		ImportantDecisions:  importantDecisions,
		Discoveries:         discoveries,
		Failures:            failures,
		Constraints:         constraints,
		UnresolvedWork:      unresolvedWork,
		RecommendedNextSteps: recommendedNextSteps,
		RelevantFiles:       relevantFiles,
		RelevantCheckpoints: relevantCheckpoints,
	}
	
	// Create Markdown content
	content := m.formatHandoff(event)
	path := fmt.Sprintf("sessions/%s/handoff.md", sessionID)
	
	if err := m.storage.WriteMarkdown(path, content); err != nil {
		return nil, fmt.Errorf("failed to write handoff: %w", err)
	}
	
	return event, nil
}

// formatHandoff formats a handoff as Markdown
func (m *Manager) formatHandoff(h *Handoff) string {
	var sb strings.Builder
	
	sb.WriteString("# Session Handoff\n\n")
	sb.WriteString(fmt.Sprintf("**Session ID:** %s\n\n", h.SessionID))
	sb.WriteString(fmt.Sprintf("**Created:** %s\n\n", h.Timestamp.Format(time.RFC3339)))
	
	sb.WriteString("## Current State\n\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", h.CurrentState))
	
	sb.WriteString("## What Changed\n\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", h.WhatChanged))
	
	if len(h.ImportantDecisions) > 0 {
		sb.WriteString("## Important Decisions\n\n")
		for _, dec := range h.ImportantDecisions {
			sb.WriteString(fmt.Sprintf("- %s\n", dec))
		}
		sb.WriteString("\n")
	}
	
	if len(h.Discoveries) > 0 {
		sb.WriteString("## Discoveries\n\n")
		for _, disc := range h.Discoveries {
			sb.WriteString(fmt.Sprintf("- %s\n", disc))
		}
		sb.WriteString("\n")
	}
	
	if len(h.Failures) > 0 {
		sb.WriteString("## Failures\n\n")
		for _, fail := range h.Failures {
			sb.WriteString(fmt.Sprintf("- %s\n", fail))
		}
		sb.WriteString("\n")
	}
	
	if len(h.Constraints) > 0 {
		sb.WriteString("## Constraints\n\n")
		for _, cons := range h.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", cons))
		}
		sb.WriteString("\n")
	}
	
	if h.UnresolvedWork != "" {
		sb.WriteString("## Unresolved Work\n\n")
		sb.WriteString(fmt.Sprintf("%s\n\n", h.UnresolvedWork))
	}
	
	if h.RecommendedNextSteps != "" {
		sb.WriteString("## Recommended Next Steps\n\n")
		sb.WriteString(fmt.Sprintf("%s\n\n", h.RecommendedNextSteps))
	}
	
	if len(h.RelevantFiles) > 0 {
		sb.WriteString("## Relevant Files\n\n")
		for _, file := range h.RelevantFiles {
			sb.WriteString(fmt.Sprintf("- %s\n", file))
		}
		sb.WriteString("\n")
	}
	
	if len(h.RelevantCheckpoints) > 0 {
		sb.WriteString("## Relevant Checkpoints\n\n")
		for _, cp := range h.RelevantCheckpoints {
			sb.WriteString(fmt.Sprintf("- %s\n", cp))
		}
		sb.WriteString("\n")
	}
	
	return sb.String()
}
