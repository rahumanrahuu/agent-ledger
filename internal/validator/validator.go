package validator

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ValidationLevel represents the severity of a validation issue
type ValidationLevel string

const (
	LevelError   ValidationLevel = "error"
	LevelWarning ValidationLevel = "warning"
	LevelInfo    ValidationLevel = "info"
)

// ValidationIssue represents a single validation problem
type ValidationIssue struct {
	Level       ValidationLevel `json:"level"`
	Category    string          `json:"category"`
	Message     string          `json:"message"`
	Path        string          `json:"path,omitempty"`
	Suggestion  string          `json:"suggestion,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
}

// ValidationResult contains results of a validation run
type ValidationResult struct {
	Valid       bool                `json:"valid"`
	Issues      []ValidationIssue   `json:"issues"`
	Warnings    int                 `json:"warning_count"`
	Errors      int                 `json:"error_count"`
	Duration    string              `json:"duration"`
	Timestamp   time.Time           `json:"timestamp"`
}

// Validator performs validation checks on ledger data
type Validator struct {
	issues []ValidationIssue
	start  time.Time
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		issues: make([]ValidationIssue, 0),
		start:  time.Now(),
	}
}

// AddIssue adds a validation issue
func (v *Validator) AddIssue(level ValidationLevel, category, message, path, suggestion string) {
	v.issues = append(v.issues, ValidationIssue{
		Level:      level,
		Category:   category,
		Message:    message,
		Path:       path,
		Suggestion: suggestion,
		Timestamp:  time.Now(),
	})
}

// ValidateSessionMetadata checks if session metadata is valid
func (v *Validator) ValidateSessionMetadata(sessionID, agent, model, branch string) {
	if sessionID == "" {
		v.AddIssue(LevelError, "session", "Session ID is empty", "", "Session ID is required")
		return
	}

	if !isValidUUID(sessionID) && !isValidID(sessionID) {
		v.AddIssue(LevelWarning, "session", "Session ID format is unusual", sessionID, "Session ID should be UUID or alphanumeric")
	}

	if agent == "" {
		v.AddIssue(LevelInfo, "session", "Agent field is empty", sessionID, "Consider setting agent field for better tracking")
	}

	if model == "" {
		v.AddIssue(LevelInfo, "session", "Model field is empty", sessionID, "Consider setting model field for analytics")
	}

	if branch == "" {
		v.AddIssue(LevelWarning, "session", "Branch field is empty", sessionID, "Branch should be recorded with session")
	}
}

// ValidateRecord checks if a record has valid structure
func (v *Validator) ValidateRecord(recordType, id, title, content, path string) {
	if title == "" {
		v.AddIssue(LevelWarning, recordType, "Record title is empty", path, "Every record should have a descriptive title")
	}

	if len(title) > 200 {
		v.AddIssue(LevelWarning, recordType, "Record title is very long", path, fmt.Sprintf("Consider shortening title from %d to <100 chars", len(title)))
	}

	if len(content) == 0 {
		v.AddIssue(LevelWarning, recordType, "Record content is empty", path, "Record should have content describing the decision/discovery/etc")
	}

	if len(content) > 50000 {
		v.AddIssue(LevelWarning, recordType, "Record content is very large", path, fmt.Sprintf("Content is %d chars, consider splitting into multiple records", len(content)))
	}

	// Accept both exact and common plural forms (e.g. "decision" and "decisions", "discovery" and "discoveries")
	validPrefix := strings.HasPrefix(path, recordType+"/") || strings.HasPrefix(path, recordType+"s/") || strings.HasPrefix(path, recordType+"ies/") || strings.HasPrefix(path, strings.TrimSuffix(recordType, "y")+"ies/")
	if !validPrefix {
		v.AddIssue(LevelWarning, recordType, "Record path doesn't match type", path, fmt.Sprintf("Record of type %s should be in %s/ or %ss/ directory", recordType, recordType, recordType))
	}
}

// ValidateDateConsistency checks if date fields are consistent
func (v *Validator) ValidateDateConsistency(startTime, endTime *time.Time, path string) {
	if startTime == nil {
		v.AddIssue(LevelError, "date", "Start time is missing", path, "Start time is required")
		return
	}

	if endTime != nil && endTime.Before(*startTime) {
		v.AddIssue(LevelError, "date", "End time is before start time", path, "Check that session end time is after start time")
	}

	if startTime.After(time.Now()) {
		v.AddIssue(LevelWarning, "date", "Start time is in the future", path, "Session start time should not be in the future")
	}

	if endTime != nil && endTime.After(time.Now().Add(24*time.Hour)) {
		v.AddIssue(LevelWarning, "date", "End time is far in the future", path, "Session end time should be around now")
	}
}

// ValidateFileFormat checks if file format is valid
func (v *Validator) ValidateFileFormat(path string, content string) {
	if !strings.HasSuffix(path, ".md") {
		v.AddIssue(LevelWarning, "format", "Non-markdown file detected", path, "Ledger records should be markdown files (.md)")
		return
	}

	// Check for basic markdown structure
	lines := strings.Split(content, "\n")
	hasHeader := false

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			hasHeader = true
			break
		}
	}

	if !hasHeader && len(content) > 0 {
		v.AddIssue(LevelInfo, "format", "Record has no markdown headers", path, "Consider starting record with # Title for better readability")
	}
}

// ValidateDuplicates checks for duplicate entries
func (v *Validator) ValidateDuplicates(records []string) {
	seen := make(map[string]int)

	for _, record := range records {
		seen[record]++
	}

	for record, count := range seen {
		if count > 1 {
			v.AddIssue(LevelWarning, "duplicates", fmt.Sprintf("Duplicate record found %d times", count), record, "Remove or consolidate duplicate entries")
		}
	}
}

// Result returns the validation result
func (v *Validator) Result() ValidationResult {
	errorCount := 0
	warningCount := 0

	for _, issue := range v.issues {
		switch issue.Level {
		case LevelError:
			errorCount++
		case LevelWarning:
			warningCount++
		}
	}

	duration := time.Since(v.start)

	return ValidationResult{
		Valid:      errorCount == 0,
		Issues:     v.issues,
		Errors:     errorCount,
		Warnings:   warningCount,
		Duration:   fmt.Sprintf("%dms", duration.Milliseconds()),
		Timestamp:  time.Now(),
	}
}

// isValidUUID checks if a string looks like a UUID
func isValidUUID(s string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(s))
}

// isValidID checks if a string is a valid ID
func isValidID(s string) bool {
	idRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return idRegex.MatchString(s) && len(s) > 0
}

// ComprehensiveValidation runs all validation checks
type ComprehensiveValidation struct {
	SessionCount    int
	RecordCount     int
	WarningCount    int
	ErrorCount      int
	Issues          []ValidationIssue
}

// RunComprehensiveValidation performs a complete validation
func RunComprehensiveValidation(sessionCount, recordCount int) ComprehensiveValidation {
	v := NewValidator()

	// Validate counts
	if sessionCount == 0 {
		v.AddIssue(LevelInfo, "stats", "No sessions found", "", "This is fine for a new project")
	}

	if recordCount == 0 && sessionCount > 0 {
		v.AddIssue(LevelInfo, "stats", "Sessions exist but no records found", "", "Consider recording decisions/discoveries from your sessions")
	}

	result := v.Result()

	return ComprehensiveValidation{
		SessionCount: sessionCount,
		RecordCount:  recordCount,
		WarningCount: result.Warnings,
		ErrorCount:   result.Errors,
		Issues:       result.Issues,
	}
}
