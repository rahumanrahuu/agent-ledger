package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Constraint represents a project constraint
type Constraint struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // CRITICAL, HIGH, MEDIUM, LOW
	AppliesTo   string    `json:"applies_to"` // glob pattern
	CreatedAt   time.Time `json:"created_at"`
	Rule        string    `json:"rule"` // markdown rule text
	Examples    []string  `json:"examples"`
}

// Violation represents a constraint violation
type Violation struct {
	ConstraintID string `json:"constraint_id"`
	File         string `json:"file"`
	Line         int    `json:"line"`
	Message      string `json:"message"`
	Severity     string `json:"severity"`
	Suggestion   string `json:"suggestion"`
}

// ConstraintChecker checks for constraint violations
type ConstraintChecker struct {
	constraints map[string]*Constraint
	root        string
}

// NewConstraintChecker creates a new constraint checker
func NewConstraintChecker(root string) (*ConstraintChecker, error) {
	cc := &ConstraintChecker{
		constraints: make(map[string]*Constraint),
		root:        root,
	}

	// Load constraints from .agent/constraints/ directory
	constraintDir := filepath.Join(root, ".agent", "constraints")
	if _, err := os.Stat(constraintDir); os.IsNotExist(err) {
		return cc, nil // No constraints yet
	}

	entries, err := os.ReadDir(constraintDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		constraintID := strings.TrimSuffix(entry.Name(), ".md")
		filePath := filepath.Join(constraintDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		constraint := parseConstraintFile(constraintID, string(data))
		cc.constraints[constraintID] = constraint
	}

	return cc, nil
}

// parseConstraintFile parses a constraint markdown file
func parseConstraintFile(id string, content string) *Constraint {
	lines := strings.Split(content, "\n")

	constraint := &Constraint{
		ID:          id,
		Rule:        content,
		CreatedAt:   time.Now(),
	}

	// Extract metadata from front matter
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			constraint.Name = strings.TrimPrefix(line, "# ")
		}
		if strings.Contains(line, "severity:") {
			constraint.Severity = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.Contains(line, "applies_to:") {
			constraint.AppliesTo = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}

	return constraint
}

// CheckFile checks a file for constraint violations
func (cc *ConstraintChecker) CheckFile(filePath string) []Violation {
	var violations []Violation

	data, err := os.ReadFile(filePath)
	if err != nil {
		return violations
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Check each constraint
	for _, constraint := range cc.constraints {
		if !matchGlob(filePath, constraint.AppliesTo) {
			continue
		}

		// Check for violations based on constraint rules
		fileViolations := checkConstraintRule(constraint, filePath, lines, content)
		violations = append(violations, fileViolations...)
	}

	return violations
}

// checkConstraintRule checks a specific constraint rule
func checkConstraintRule(constraint *Constraint, filePath string, lines []string, content string) []Violation {
	var violations []Violation

	// Example: auth-otp-only constraint
	if constraint.ID == "auth-otp-only" {
		if strings.Contains(content, "firebase") || strings.Contains(content, "Firebase") {
			violations = append(violations, Violation{
				ConstraintID: constraint.ID,
				File:         filePath,
				Message:      "Firebase detected - must use Supabase OTP only",
				Severity:     constraint.Severity,
				Suggestion:   "Replace Firebase with Supabase auth.signInWithOtp()",
			})
		}
	}

	// Example: rate-limit constraint
	if constraint.ID == "rate-limit" {
		// Check for hardcoded rate limits that don't match constraint
		for _, line := range lines {
			if strings.Contains(line, "8") && strings.Contains(line, "day") {
				continue // This is correct
			}
		}
	}

	return violations
}

// matchGlob checks if a file matches a glob pattern
func matchGlob(file string, pattern string) bool {
	pattern = strings.ReplaceAll(pattern, "**", ".*")
	pattern = strings.ReplaceAll(pattern, "*", "[^/]*")
	regex := regexp.MustCompile("^" + pattern + "$")
	return regex.MatchString(file)
}

// ListConstraints returns all active constraints
func (cc *ConstraintChecker) ListConstraints() []*Constraint {
	var constraints []*Constraint
	for _, c := range cc.constraints {
		constraints = append(constraints, c)
	}
	return constraints
}

// Format checks for display
func (violation Violation) String() string {
	return fmt.Sprintf(
		"🔴 %s\n   File: %s\n   Severity: %s\n   Message: %s\n   Suggestion: %s",
		violation.ConstraintID,
		violation.File,
		violation.Severity,
		violation.Message,
		violation.Suggestion,
	)
}
