package events

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"agent-ledger/internal/storage"
)

// parseTimestamp parses the timestamp from a markdown line, stripping
// the leading "*Created:" prefix and any trailing "*" characters that
// the markdown formatter appends. Returns zero time if parsing fails.
func parseTimestamp(line string) time.Time {
	// Strip leading prefix
	s := strings.TrimPrefix(line, "*Created:")
	// Strip any surrounding whitespace and trailing asterisks
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "*")
	s = strings.TrimSpace(s)
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// parseSessionID parses the session ID from a markdown line.
func parseSessionID(line string) string {
	s := strings.TrimPrefix(line, "*Session:")
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "*")
	return strings.TrimSpace(s)
}

// parseAgent parses the agent name from a markdown line.
func parseAgent(line string) string {
	s := strings.TrimPrefix(line, "*Agent:")
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "*")
	return strings.TrimSpace(s)
}

// ParseDecision parses a Decision from markdown content.
// Returns nil if the content is not a valid decision record.
func ParseDecision(content string) *Decision {
	d := &Decision{}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "# "):
			d.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
		case strings.HasPrefix(line, "**Decision:**"):
			d.Decision = strings.TrimSpace(strings.TrimPrefix(line, "**Decision:**"))
		case strings.HasPrefix(line, "**Rationale:**"):
			d.Rationale = strings.TrimSpace(strings.TrimPrefix(line, "**Rationale:**"))
		case strings.HasPrefix(line, "*Session:"):
			d.SessionID = parseSessionID(line)
		case strings.HasPrefix(line, "*Created:"):
			d.Timestamp = parseTimestamp(line)
		case strings.HasPrefix(line, "*Agent:"):
			d.Agent = parseAgent(line)
		}
	}
	if d.Title == "" {
		return nil
	}
	return d
}

// ParseDiscovery parses a Discovery from markdown content.
// Returns nil if the content is not a valid discovery record.
func ParseDiscovery(content string) *Discovery {
	d := &Discovery{}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "# "):
			d.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
		case strings.HasPrefix(line, "**Finding:**"):
			d.Finding = strings.TrimSpace(strings.TrimPrefix(line, "**Finding:**"))
		case strings.HasPrefix(line, "*Session:"):
			d.SessionID = parseSessionID(line)
		case strings.HasPrefix(line, "*Created:"):
			d.Timestamp = parseTimestamp(line)
		case strings.HasPrefix(line, "*Agent:"):
			d.Agent = parseAgent(line)
		}
	}
	if d.Title == "" {
		return nil
	}
	return d
}

// ParseFailure parses a Failure from markdown content.
// Returns nil if the content is not a valid failure record.
func ParseFailure(content string) *Failure {
	f := &Failure{}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "# "):
			f.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
		case strings.HasPrefix(line, "**Attempted Approach:**"):
			f.AttemptedApproach = strings.TrimSpace(strings.TrimPrefix(line, "**Attempted Approach:**"))
		case strings.HasPrefix(line, "**Why It Failed:**"):
			f.WhyItFailed = strings.TrimSpace(strings.TrimPrefix(line, "**Why It Failed:**"))
		case strings.HasPrefix(line, "**Lessons:**"):
			f.Lessons = strings.TrimSpace(strings.TrimPrefix(line, "**Lessons:**"))
		case strings.HasPrefix(line, "*Session:"):
			f.SessionID = parseSessionID(line)
		case strings.HasPrefix(line, "*Created:"):
			f.Timestamp = parseTimestamp(line)
		case strings.HasPrefix(line, "*Agent:"):
			f.Agent = parseAgent(line)
		}
	}
	if f.Title == "" {
		return nil
	}
	return f
}

// ParseConstraint parses a Constraint from markdown content.
// Returns nil if the content is not a valid constraint record.
func ParseConstraint(content string) *Constraint {
	c := &Constraint{}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "# "):
			c.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
		case strings.HasPrefix(line, "**Constraint:**"):
			c.Constraint = strings.TrimSpace(strings.TrimPrefix(line, "**Constraint:**"))
		case strings.HasPrefix(line, "**Reason:**"):
			c.Reason = strings.TrimSpace(strings.TrimPrefix(line, "**Reason:**"))
		case strings.HasPrefix(line, "*Session:"):
			c.SessionID = parseSessionID(line)
		case strings.HasPrefix(line, "*Created:"):
			c.Timestamp = parseTimestamp(line)
		case strings.HasPrefix(line, "*Agent:"):
			c.Agent = parseAgent(line)
		}
	}
	if c.Title == "" {
		return nil
	}
	return c
}

// ListDecisions returns all decision records from storage, sorted newest-first.
// The count of returned records matches the number of readable .md files in decisions/.
func ListDecisions(st *storage.Storage) ([]*Decision, error) {
	files, err := st.ListFiles(storage.DecisionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list decisions: %w", err)
	}

	var results []*Decision
	for _, file := range files {
		content, err := st.ReadMarkdown(storage.DecisionsDir + "/" + file)
		if err != nil {
			continue
		}
		if d := ParseDecision(content); d != nil {
			results = append(results, d)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Timestamp.Equal(results[j].Timestamp) {
			return results[i].Title > results[j].Title // Reverse alphabetical
		}
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	return results, nil
}

// ListDiscoveries returns all discovery records from storage, sorted newest-first.
func ListDiscoveries(st *storage.Storage) ([]*Discovery, error) {
	files, err := st.ListFiles(storage.DiscoveriesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list discoveries: %w", err)
	}

	var results []*Discovery
	for _, file := range files {
		content, err := st.ReadMarkdown(storage.DiscoveriesDir + "/" + file)
		if err != nil {
			continue
		}
		if d := ParseDiscovery(content); d != nil {
			results = append(results, d)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	return results, nil
}

// ListFailures returns all failure records from storage, sorted newest-first.
func ListFailures(st *storage.Storage) ([]*Failure, error) {
	files, err := st.ListFiles(storage.FailuresDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list failures: %w", err)
	}

	var results []*Failure
	for _, file := range files {
		content, err := st.ReadMarkdown(storage.FailuresDir + "/" + file)
		if err != nil {
			continue
		}
		if f := ParseFailure(content); f != nil {
			results = append(results, f)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	return results, nil
}

// ListConstraints returns all constraint records from storage, sorted newest-first.
func ListConstraints(st *storage.Storage) ([]*Constraint, error) {
	files, err := st.ListFiles(storage.ConstraintsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list constraints: %w", err)
	}

	var results []*Constraint
	for _, file := range files {
		content, err := st.ReadMarkdown(storage.ConstraintsDir + "/" + file)
		if err != nil {
			continue
		}
		if c := ParseConstraint(content); c != nil {
			results = append(results, c)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	return results, nil
}

// CountDecisions returns the number of readable decision records.
// This count is always derived from actual files.
func CountDecisions(st *storage.Storage) int {
	decisions, _ := ListDecisions(st)
	return len(decisions)
}

// CountDiscoveries returns the number of readable discovery records.
func CountDiscoveries(st *storage.Storage) int {
	discoveries, _ := ListDiscoveries(st)
	return len(discoveries)
}

// CountFailures returns the number of readable failure records.
func CountFailures(st *storage.Storage) int {
	failures, _ := ListFailures(st)
	return len(failures)
}

// CountConstraints returns the number of readable constraint records.
func CountConstraints(st *storage.Storage) int {
	constraints, _ := ListConstraints(st)
	return len(constraints)
}
