package context

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agent-ledger/internal/checkpoint"
	"agent-ledger/internal/events"
	"agent-ledger/internal/git"
	"agent-ledger/internal/history"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
)

// Context represents compiled project context, partitioned into three clear categories.
type Context struct {
	// Facts — objective information derived from Git/filesystem/project inspection.
	Project          string
	CurrentState     string
	Architecture     string
	RecentDevelopment string
	TestingStatus    string
	RecentlyChanged  []FileChange

	// RecordedKnowledge — decisions, discoveries, failures, constraints, handoffs
	// explicitly recorded by agents. These are never generated automatically.
	ImportantDecisions []DecisionSummary
	Discoveries        []DiscoverySummary
	KnownFailures      []FailureSummary
	Constraints        []ConstraintSummary
	LatestHandoff      string

	// Recommendations — explicit recommendations made by agents or derived from
	// recorded failures/handoffs. Never generated from raw Git state alone.
	Recommendations []string

	// TaskContext — additional context when a task is specified.
	TaskContext string
}

// DecisionSummary represents a condensed decision for context.
type DecisionSummary struct {
	Title     string
	Decision  string
	Rationale string
	SessionID string
	Timestamp time.Time
	Agent     string
	IsRecent  bool
}

// DiscoverySummary represents a condensed discovery for context.
type DiscoverySummary struct {
	Title     string
	Finding   string
	SessionID string
	Timestamp time.Time
	Agent     string
	IsRecent  bool
}

// FailureSummary represents a condensed failure for context.
type FailureSummary struct {
	Title             string
	AttemptedApproach string
	WhyItFailed       string
	Lessons           string
	SessionID         string
	Timestamp         time.Time
	Agent             string
	IsRecent          bool
}

// ConstraintSummary represents a condensed constraint for context.
type ConstraintSummary struct {
	Title     string
	Constraint string
	Reason    string
	SessionID string
	Timestamp time.Time
	Agent     string
	IsRecent  bool
}

// FileChange represents a file that has been modified.
type FileChange struct {
	Path       string
	ChangeType string // "modified (staged)", "modified (unstaged)", "added (untracked)"
}

// Manager manages context compilation.
type Manager struct {
	historyManager    *history.Manager
	checkpointManager *checkpoint.Manager
	storage           *storage.Storage
}

// NewManager creates a new context manager.
func NewManager(historyManager *history.Manager, checkpointManager *checkpoint.Manager, st *storage.Storage) *Manager {
	return &Manager{
		historyManager:    historyManager,
		checkpointManager: checkpointManager,
		storage:           st,
	}
}

// recentThreshold is the boundary for marking records as "recent".
const recentThreshold = 7 * 24 * time.Hour

// Compile compiles project context for a new agent.
func (m *Manager) Compile(repo *repository.Repository, task string) (*Context, error) {
	now := time.Now()

	ctx := &Context{
		Project:           m.getProjectInfo(repo),
		CurrentState:      m.formatCurrentState(repo),
		Architecture:      m.getArchitecture(),
		RecentDevelopment: m.getRecentDevelopment(repo),
		TestingStatus:     m.getTestingStatus(),
		RecentlyChanged:   m.getRecentChanges(repo),
	}

	// Recorded knowledge — all records, sorted by recency, optionally task-filtered.
	ctx.ImportantDecisions = m.getDecisions(now, task)
	ctx.Discoveries = m.getDiscoveries(now, task)
	ctx.KnownFailures = m.getFailures(now, task)
	ctx.Constraints = m.getConstraints(now, task)
	ctx.LatestHandoff = m.getLatestHandoff()

	// Recommendations — only from explicit agent-recorded sources.
	ctx.Recommendations = m.getRecommendations()

	// Task-specific context.
	if task != "" {
		ctx.TaskContext = m.getTaskContext(task)
	}

	return ctx, nil
}

// getProjectInfo gets basic project information (FACT).
func (m *Manager) getProjectInfo(repo *repository.Repository) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Repository: %s\n", repo.Root))
	sb.WriteString(fmt.Sprintf("Branch: %s\n", repo.Branch))
	sb.WriteString(fmt.Sprintf("Head: %s\n", repo.Head))

	if len(repo.Remotes) > 0 {
		sb.WriteString("Remotes: ")
		var remoteNames []string
		for name := range repo.Remotes {
			remoteNames = append(remoteNames, name)
		}
		sort.Strings(remoteNames)
		for i, name := range remoteNames {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(name)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatCurrentState formats the current repository state (FACT).
// This section is purely descriptive — it does NOT contain recommendations.
func (m *Manager) formatCurrentState(repo *repository.Repository) string {
	var sb strings.Builder

	if repo.Dirty {
		sb.WriteString("Repository has uncommitted changes:\n")

		if len(repo.Staged) > 0 {
			sb.WriteString("  Staged changes:\n")
			for _, file := range repo.Staged {
				sb.WriteString(fmt.Sprintf("    - %s\n", file))
			}
		}

		if len(repo.Unstaged) > 0 {
			sb.WriteString("  Unstaged changes:\n")
			for _, file := range repo.Unstaged {
				sb.WriteString(fmt.Sprintf("    - %s\n", file))
			}
		}

		if len(repo.Untracked) > 0 {
			sb.WriteString("  Untracked files:\n")
			for _, file := range repo.Untracked {
				sb.WriteString(fmt.Sprintf("    - %s\n", file))
			}
		}
	} else {
		sb.WriteString("Repository is clean (no uncommitted changes)\n")
	}

	// Current session state (FACT — not a recommendation).
	sessions, err := m.storage.ListDirectories(storage.SessionsDir)
	if err == nil {
		var activeCount int
		for _, sessionID := range sessions {
			metadataPath := fmt.Sprintf("sessions/%s/metadata.json", sessionID)
			var meta struct {
				ID        string     `json:"id"`
				Agent     string     `json:"agent,omitempty"`
				Model     string     `json:"model,omitempty"`
				StartTime time.Time  `json:"start_time"`
				EndTime   *time.Time `json:"end_time,omitempty"`
			}
			if err := m.storage.ReadJSON(metadataPath, &meta); err == nil {
				if meta.EndTime == nil {
					activeCount++
					sb.WriteString(fmt.Sprintf("Active session: %s", meta.ID))
					if meta.Agent != "" {
						sb.WriteString(fmt.Sprintf(" (agent: %s", meta.Agent))
						if meta.Model != "" {
							sb.WriteString(fmt.Sprintf(", model: %s", meta.Model))
						}
						sb.WriteString(")")
					}
					sb.WriteString(fmt.Sprintf(", started %s\n", meta.StartTime.Format("2006-01-02 15:04:05 UTC")))
				}
			}
		}
		if activeCount == 0 {
			sb.WriteString("No active session\n")
		}
	}

	return sb.String()
}

// getArchitecture analyzes the project structure (FACT).
func (m *Manager) getArchitecture() string {
	var sb strings.Builder

	repoRoot, err := repository.GetRepositoryRoot()
	if err != nil || repoRoot == "" {
		return "Unable to determine architecture - not in a git repository"
	}

	sb.WriteString("Package Structure:\n")

	// Analyze cmd directory.
	cmdDir := filepath.Join(repoRoot, "cmd")
	if dirs, err := os.ReadDir(cmdDir); err == nil {
		for _, dir := range dirs {
			if dir.IsDir() {
				sb.WriteString(fmt.Sprintf("  cmd/%s/ - CLI entry point\n", dir.Name()))
				mainFile := filepath.Join(cmdDir, dir.Name(), "main.go")
				if _, err := os.Stat(mainFile); err == nil {
					sb.WriteString(fmt.Sprintf("    → %s\n", mainFile))
				}
			}
		}
	}

	// Analyze internal directory.
	internalDir := filepath.Join(repoRoot, "internal")
	if dirs, err := os.ReadDir(internalDir); err == nil {
		for _, dir := range dirs {
			if dir.IsDir() {
				pkgDir := filepath.Join(internalDir, dir.Name())
				var goFiles, testFiles int
				if files, err := os.ReadDir(pkgDir); err == nil {
					for _, file := range files {
						if strings.HasSuffix(file.Name(), "_test.go") {
							testFiles++
						} else if strings.HasSuffix(file.Name(), ".go") {
							goFiles++
						}
					}
				}
				label := "Core package"
				sb.WriteString(fmt.Sprintf("  internal/%s/ - %s\n", dir.Name(), label))
				if goFiles > 0 {
					sb.WriteString(fmt.Sprintf("    → %d Go source file(s)", goFiles))
					if testFiles > 0 {
						sb.WriteString(fmt.Sprintf(", %d test file(s)", testFiles))
					}
					sb.WriteString("\n")
				}
			}
		}
	}

	// Analyze mcp directory.
	mcpDir := filepath.Join(repoRoot, "mcp")
	if files, err := os.ReadDir(mcpDir); err == nil {
		var mcpSrc, mcpTest int
		for _, file := range files {
			if strings.HasSuffix(file.Name(), "_test.go") {
				mcpTest++
			} else if strings.HasSuffix(file.Name(), ".go") {
				mcpSrc++
			}
		}
		if mcpSrc > 0 {
			sb.WriteString(fmt.Sprintf("  mcp/ - MCP server (%d source file(s)", mcpSrc))
			if mcpTest > 0 {
				sb.WriteString(fmt.Sprintf(", %d test file(s)", mcpTest))
			}
			sb.WriteString(")\n")
		}
	}

	sb.WriteString("\nCore Dependencies:\n")
	sb.WriteString("  CLI → Core services (git, session, checkpoint, context, history, events, storage)\n")
	sb.WriteString("  MCP → Same core services\n")

	return sb.String()
}

// getRecentDevelopment gets recent development information (FACT).
func (m *Manager) getRecentDevelopment(repo *repository.Repository) string {
	var sb strings.Builder

	logOutput, err := git.Command("log", "--oneline", "--no-decorate", "-10")
	if err != nil {
		return "Unable to retrieve recent commits"
	}

	commits := strings.Split(strings.TrimSpace(logOutput), "\n")
	if len(commits) > 0 && commits[0] != "" {
		sb.WriteString("Recent Commits:\n")
		for _, commit := range commits {
			if commit == "" {
				continue
			}
			sb.WriteString(fmt.Sprintf("  %s\n", commit))
		}
	}

	// Recent changes by package.
	packages := map[string]int{}
	for _, file := range repo.Staged {
		if pkg := extractPackage(file); pkg != "" {
			packages[pkg]++
		}
	}
	for _, file := range repo.Unstaged {
		if pkg := extractPackage(file); pkg != "" {
			packages[pkg]++
		}
	}

	if len(packages) > 0 {
		sb.WriteString("\nRecent Changes by Package:\n")
		var pkgNames []string
		for pkg := range packages {
			pkgNames = append(pkgNames, pkg)
		}
		sort.Strings(pkgNames)
		for _, pkg := range pkgNames {
			sb.WriteString(fmt.Sprintf("  %s: %d file(s)\n", pkg, packages[pkg]))
		}
	}

	return sb.String()
}

// extractPackage extracts the internal package name from a file path.
func extractPackage(filePath string) string {
	if strings.Contains(filePath, "internal/") {
		parts := strings.Split(filePath, "/")
		for i, part := range parts {
			if part == "internal" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	return ""
}

// getDecisions returns all decisions, sorted newest-first.
// NEVER silently drops records — all parseable decisions are returned.
// When task is provided, results are additionally filtered by relevance.
func (m *Manager) getDecisions(now time.Time, task string) []DecisionSummary {
	all, err := events.ListDecisions(m.storage)
	if err != nil {
		return nil
	}

	var summaries []DecisionSummary
	for _, d := range all {
		if task != "" && !isRelevantToTask(d.Decision+" "+d.Rationale+" "+d.Title, task) {
			continue
		}
		summaries = append(summaries, DecisionSummary{
			Title:     d.Title,
			Decision:  d.Decision,
			Rationale: d.Rationale,
			SessionID: d.SessionID,
			Timestamp: d.Timestamp,
			Agent:     d.Agent,
			IsRecent:  now.Sub(d.Timestamp) < recentThreshold,
		})
	}

	// Sort: recent first, then alphabetical by title for stability.
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].IsRecent != summaries[j].IsRecent {
			return summaries[i].IsRecent
		}
		return summaries[i].Timestamp.After(summaries[j].Timestamp)
	})

	// Cap at 10; architectural decisions are always included.
	if len(summaries) > 10 {
		summaries = summaries[:10]
	}
	return summaries
}

// getDiscoveries returns all discoveries, sorted newest-first.
// NEVER silently drops records.
func (m *Manager) getDiscoveries(now time.Time, task string) []DiscoverySummary {
	all, err := events.ListDiscoveries(m.storage)
	if err != nil {
		return nil
	}

	var summaries []DiscoverySummary
	for _, d := range all {
		if task != "" && !isRelevantToTask(d.Finding+" "+d.Title, task) {
			continue
		}
		summaries = append(summaries, DiscoverySummary{
			Title:     d.Title,
			Finding:   d.Finding,
			SessionID: d.SessionID,
			Timestamp: d.Timestamp,
			Agent:     d.Agent,
			IsRecent:  now.Sub(d.Timestamp) < recentThreshold,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Timestamp.After(summaries[j].Timestamp)
	})

	if len(summaries) > 6 {
		summaries = summaries[:6]
	}
	return summaries
}

// getFailures returns all failures, sorted newest-first.
// NEVER silently drops records.
func (m *Manager) getFailures(now time.Time, task string) []FailureSummary {
	all, err := events.ListFailures(m.storage)
	if err != nil {
		return nil
	}

	var summaries []FailureSummary
	for _, f := range all {
		if task != "" && !isRelevantToTask(f.AttemptedApproach+" "+f.WhyItFailed+" "+f.Title, task) {
			continue
		}
		summaries = append(summaries, FailureSummary{
			Title:             f.Title,
			AttemptedApproach: f.AttemptedApproach,
			WhyItFailed:       f.WhyItFailed,
			Lessons:           f.Lessons,
			SessionID:         f.SessionID,
			Timestamp:         f.Timestamp,
			Agent:             f.Agent,
			IsRecent:          now.Sub(f.Timestamp) < recentThreshold,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Timestamp.After(summaries[j].Timestamp)
	})

	if len(summaries) > 5 {
		summaries = summaries[:5]
	}
	return summaries
}

// getConstraints returns all constraints, sorted newest-first.
// Constraints have no expiry — they represent permanent project invariants.
func (m *Manager) getConstraints(now time.Time, task string) []ConstraintSummary {
	all, err := events.ListConstraints(m.storage)
	if err != nil {
		return nil
	}

	var summaries []ConstraintSummary
	for _, c := range all {
		if task != "" && !isRelevantToTask(c.Constraint+" "+c.Reason+" "+c.Title, task) {
			continue
		}
		summaries = append(summaries, ConstraintSummary{
			Title:     c.Title,
			Constraint: c.Constraint,
			Reason:    c.Reason,
			SessionID: c.SessionID,
			Timestamp: c.Timestamp,
			Agent:     c.Agent,
			IsRecent:  now.Sub(c.Timestamp) < recentThreshold,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Timestamp.After(summaries[j].Timestamp)
	})

	return summaries
}

// getTestingStatus analyzes test coverage (FACT).
func (m *Manager) getTestingStatus() string {
	var sb strings.Builder

	repoRoot, err := repository.GetRepositoryRoot()
	if err != nil || repoRoot == "" {
		return "Unable to determine testing status"
	}

	packages := []string{"storage", "session", "git", "checkpoint", "context", "events", "history", "repository"}

	var withTests, withoutTests []string
	for _, pkg := range packages {
		testFile := filepath.Join(repoRoot, "internal", pkg, pkg+"_test.go")
		if _, err := os.Stat(testFile); err == nil {
			withTests = append(withTests, pkg)
		} else {
			withoutTests = append(withoutTests, pkg)
		}
	}

	// Check mcp package.
	mcpTestFile := filepath.Join(repoRoot, "mcp", "mcp_test.go")
	if _, err := os.Stat(mcpTestFile); err == nil {
		withTests = append(withTests, "mcp")
	} else {
		withoutTests = append(withoutTests, "mcp")
	}

	if len(withTests) > 0 {
		sb.WriteString("Test Coverage:\n")
		for _, pkg := range withTests {
			if pkg == "mcp" {
				sb.WriteString(fmt.Sprintf("  mcp: has tests\n"))
			} else {
				sb.WriteString(fmt.Sprintf("  internal/%s: has tests\n", pkg))
			}
		}
	}

	if len(withoutTests) > 0 {
		sb.WriteString("Missing Tests:\n")
		for _, pkg := range withoutTests {
			sb.WriteString(fmt.Sprintf("  internal/%s: no tests\n", pkg))
		}
	}

	return sb.String()
}

// getRecentChanges gets meaningful file changes (FACT).
func (m *Manager) getRecentChanges(repo *repository.Repository) []FileChange {
	var changes []FileChange

	for _, file := range repo.Staged {
		changes = append(changes, FileChange{Path: file, ChangeType: "modified (staged)"})
	}
	for _, file := range repo.Unstaged {
		changes = append(changes, FileChange{Path: file, ChangeType: "modified (unstaged)"})
	}
	for _, file := range repo.Untracked {
		if !strings.HasPrefix(file, ".agent/") {
			changes = append(changes, FileChange{Path: file, ChangeType: "added (untracked)"})
		}
	}

	if len(changes) > 15 {
		changes = changes[:15]
	}
	return changes
}

// getLatestHandoff returns the content of the most recent session handoff.
// Includes sessions that have no EndTime (sessions with handoffs but never formally stopped).
func (m *Manager) getLatestHandoff() string {
	sessions, err := m.storage.ListDirectories(storage.SessionsDir)
	if err != nil {
		return ""
	}

	var latestHandoff string
	var latestTime time.Time

	for _, sessionID := range sessions {
		handoffPath := fmt.Sprintf("sessions/%s/handoff.md", sessionID)
		if !m.storage.FileExists(handoffPath) {
			continue
		}

		metadataPath := fmt.Sprintf("sessions/%s/metadata.json", sessionID)
		var meta struct {
			StartTime time.Time  `json:"start_time"`
			EndTime   *time.Time `json:"end_time,omitempty"`
		}
		if err := m.storage.ReadJSON(metadataPath, &meta); err != nil {
			continue
		}

		// Use EndTime if available; fall back to StartTime.
		// This ensures sessions with handoffs but no EndTime are included.
		sessionTime := meta.StartTime
		if meta.EndTime != nil {
			sessionTime = *meta.EndTime
		}

		if sessionTime.After(latestTime) {
			latestTime = sessionTime
			content, err := m.storage.ReadMarkdown(handoffPath)
			if err == nil {
				latestHandoff = content
			}
		}
	}

	return latestHandoff
}

// getRecommendations returns explicit recommendations from recorded sources only.
// Does NOT generate recommendations from raw Git state.
func (m *Manager) getRecommendations() []string {
	var recs []string

	// Only source: failures that explicitly record a lesson.
	failures, _ := events.ListFailures(m.storage)
	for _, f := range failures {
		if f.Lessons != "" {
			recs = append(recs, fmt.Sprintf("Lesson from failure '%s': %s", f.Title, f.Lessons))
		}
	}

	return recs
}

// getTaskContext returns task-specific context using keyword and package matching.
func (m *Manager) getTaskContext(task string) string {
	if task == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Task: %s\n\n", task))
	taskLower := strings.ToLower(task)

	// Relevant packages.
	packageKeywords := map[string][]string{
		"storage":    {"storage", "file", "disk", "write", "read", "persist"},
		"session":    {"session", "start", "stop", "agent"},
		"checkpoint": {"checkpoint", "git", "commit", "tree", "index", "snapshot"},
		"context":    {"context", "compile", "format", "summary"},
		"git":        {"git", "branch", "head", "commit", "ref", "tree"},
		"history":    {"history", "past", "previous", "session"},
		"events":     {"decision", "discovery", "failure", "constraint", "event", "record"},
		"mcp":        {"mcp", "tool", "resource", "server"},
	}

	var relevantPkgs []string
	for pkg, keywords := range packageKeywords {
		for _, kw := range keywords {
			if strings.Contains(taskLower, kw) {
				relevantPkgs = append(relevantPkgs, pkg)
				break
			}
		}
	}
	sort.Strings(relevantPkgs)

	if len(relevantPkgs) > 0 {
		sb.WriteString("Relevant packages:\n")
		for _, pkg := range relevantPkgs {
			if pkg == "mcp" {
				sb.WriteString("  mcp/\n")
			} else {
				sb.WriteString(fmt.Sprintf("  internal/%s/\n", pkg))
			}
		}
		sb.WriteString("\n")
	}

	// Relevant decisions.
	decisions, _ := events.ListDecisions(m.storage)
	var relevantDecisions []string
	for _, d := range decisions {
		if isRelevantToTask(d.Title+" "+d.Decision+" "+d.Rationale, task) {
			relevantDecisions = append(relevantDecisions, fmt.Sprintf("- %s: %s", d.Title, d.Decision))
		}
	}
	if len(relevantDecisions) > 0 {
		sb.WriteString("Relevant decisions:\n")
		for _, d := range relevantDecisions {
			sb.WriteString("  " + d + "\n")
		}
		sb.WriteString("\n")
	}

	// Relevant discoveries.
	discoveries, _ := events.ListDiscoveries(m.storage)
	var relevantDiscoveries []string
	for _, d := range discoveries {
		if isRelevantToTask(d.Title+" "+d.Finding, task) {
			relevantDiscoveries = append(relevantDiscoveries, fmt.Sprintf("- %s: %s", d.Title, d.Finding))
		}
	}
	if len(relevantDiscoveries) > 0 {
		sb.WriteString("Relevant discoveries:\n")
		for _, d := range relevantDiscoveries {
			sb.WriteString("  " + d + "\n")
		}
		sb.WriteString("\n")
	}

	// Relevant failures.
	failures, _ := events.ListFailures(m.storage)
	var relevantFailures []string
	for _, f := range failures {
		if isRelevantToTask(f.Title+" "+f.AttemptedApproach+" "+f.WhyItFailed, task) {
			relevantFailures = append(relevantFailures, fmt.Sprintf("- %s: %s", f.Title, f.WhyItFailed))
		}
	}
	if len(relevantFailures) > 0 {
		sb.WriteString("Relevant failures:\n")
		for _, f := range relevantFailures {
			sb.WriteString("  " + f + "\n")
		}
	}

	return sb.String()
}

// isRelevantToTask checks if text is relevant to a task using simple keyword matching.
func isRelevantToTask(text, task string) bool {
	textLower := strings.ToLower(text)
	// Split task into words and check each.
	words := strings.Fields(strings.ToLower(task))
	// Require at least one significant word to match.
	significantWords := 0
	matches := 0
	for _, word := range words {
		if len(word) < 4 {
			continue // skip short words
		}
		significantWords++
		if strings.Contains(textLower, word) {
			matches++
		}
	}
	if significantWords == 0 {
		return false
	}
	return matches > 0
}

// Format formats the context as a readable string with clear section separation.
//
// Section structure:
//
//	FACTS           — objective Git/filesystem observations
//	RECORDED KNOWLEDGE — decisions, discoveries, failures, constraints, handoffs
//	RECOMMENDATIONS — explicit lessons from failures / agent handoff recommendations
func (m *Manager) Format(ctx *Context) string {
	var sb strings.Builder

	sb.WriteString("PROJECT CONTEXT\n\n")

	// ─── FACTS ───────────────────────────────────────────────────────────────
	sb.WriteString("═══ FACTS ═══\n")
	sb.WriteString("(Objective information derived from Git/filesystem inspection)\n\n")

	sb.WriteString("PROJECT\n")
	sb.WriteString(ctx.Project + "\n")

	sb.WriteString("CURRENT STATE\n")
	sb.WriteString(ctx.CurrentState + "\n")

	if ctx.Architecture != "" {
		sb.WriteString("ARCHITECTURE\n")
		sb.WriteString(ctx.Architecture + "\n")
	}

	if ctx.RecentDevelopment != "" {
		sb.WriteString("RECENT DEVELOPMENT\n")
		sb.WriteString(ctx.RecentDevelopment + "\n")
	}

	if ctx.TestingStatus != "" {
		sb.WriteString("TESTING STATUS\n")
		sb.WriteString(ctx.TestingStatus + "\n")
	}

	if len(ctx.RecentlyChanged) > 0 {
		sb.WriteString("RECENTLY CHANGED AREAS\n")
		for _, change := range ctx.RecentlyChanged {
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", change.Path, change.ChangeType))
		}
		sb.WriteString("\n")
	}

	// ─── RECORDED KNOWLEDGE ──────────────────────────────────────────────────
	sb.WriteString("═══ RECORDED KNOWLEDGE ═══\n")
	sb.WriteString("(Decisions, discoveries, failures, and constraints explicitly recorded by agents)\n\n")

	if len(ctx.ImportantDecisions) > 0 {
		sb.WriteString(fmt.Sprintf("DECISIONS (%d)\n", len(ctx.ImportantDecisions)))
		for _, decision := range ctx.ImportantDecisions {
			sb.WriteString(fmt.Sprintf("- %s", decision.Title))
			if decision.IsRecent {
				sb.WriteString(" [recent]")
			}
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("  Decision: %s\n", decision.Decision))
			if decision.Rationale != "" {
				sb.WriteString(fmt.Sprintf("  Rationale: %s\n", decision.Rationale))
			}
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("DECISIONS (0)\n  No decisions recorded yet.\n\n")
	}

	if len(ctx.Discoveries) > 0 {
		sb.WriteString(fmt.Sprintf("DISCOVERIES (%d)\n", len(ctx.Discoveries)))
		for _, discovery := range ctx.Discoveries {
			sb.WriteString(fmt.Sprintf("- %s", discovery.Title))
			if discovery.IsRecent {
				sb.WriteString(" [recent]")
			}
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("  Finding: %s\n", discovery.Finding))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("DISCOVERIES (0)\n  No discoveries recorded yet.\n\n")
	}

	if len(ctx.KnownFailures) > 0 {
		sb.WriteString(fmt.Sprintf("KNOWN FAILURES (%d)\n", len(ctx.KnownFailures)))
		for _, failure := range ctx.KnownFailures {
			sb.WriteString(fmt.Sprintf("- %s", failure.Title))
			if failure.IsRecent {
				sb.WriteString(" [recent]")
			}
			sb.WriteString("\n")
			if failure.WhyItFailed != "" {
				sb.WriteString(fmt.Sprintf("  Why: %s\n", failure.WhyItFailed))
			}
			if failure.Lessons != "" {
				sb.WriteString(fmt.Sprintf("  Lessons: %s\n", failure.Lessons))
			}
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("KNOWN FAILURES (0)\n  No failures recorded yet.\n\n")
	}

	if len(ctx.Constraints) > 0 {
		sb.WriteString(fmt.Sprintf("CONSTRAINTS (%d)\n", len(ctx.Constraints)))
		for _, constraint := range ctx.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", constraint.Title))
			sb.WriteString(fmt.Sprintf("  Constraint: %s\n", constraint.Constraint))
			if constraint.Reason != "" {
				sb.WriteString(fmt.Sprintf("  Reason: %s\n", constraint.Reason))
			}
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("CONSTRAINTS (0)\n  No constraints recorded yet.\n\n")
	}

	if ctx.LatestHandoff != "" {
		sb.WriteString("LATEST HANDOFF\n")
		sb.WriteString(ctx.LatestHandoff)
		sb.WriteString("\n")
	} else {
		sb.WriteString("LATEST HANDOFF\n  No handoff recorded yet.\n\n")
	}

	// ─── RECOMMENDATIONS ─────────────────────────────────────────────────────
	// Only emit this section when there are explicit agent-recorded recommendations.
	if len(ctx.Recommendations) > 0 {
		sb.WriteString("═══ RECOMMENDATIONS ═══\n")
		sb.WriteString("(Explicit recommendations derived from agent-recorded failures/handoffs only)\n\n")
		for _, rec := range ctx.Recommendations {
			sb.WriteString(fmt.Sprintf("- %s\n", rec))
		}
		sb.WriteString("\n")
	}

	// ─── TASK-SPECIFIC ───────────────────────────────────────────────────────
	if ctx.TaskContext != "" {
		sb.WriteString("═══ TASK-SPECIFIC CONTEXT ═══\n")
		sb.WriteString(ctx.TaskContext)
		sb.WriteString("\n")
	}

	return sb.String()
}

// Explain explains changes to a file without a session context.
func (m *Manager) Explain(filePath string) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("FILE: %s\n\n", filePath))

	// Git history for the file.
	logOutput, err := git.Command("log", "--oneline", "--", filePath)
	if err != nil || strings.TrimSpace(logOutput) == "" {
		sb.WriteString("GIT HISTORY\n  No committed history found for this file.\n\n")
	} else {
		sb.WriteString("GIT HISTORY\n")
		for _, line := range strings.Split(strings.TrimSpace(logOutput), "\n") {
			sb.WriteString(fmt.Sprintf("  %s\n", line))
		}
		sb.WriteString("\n")
	}

	// Package context.
	sb.WriteString("PACKAGE CONTEXT\n")
	repoRoot, _ := repository.GetRepositoryRoot()
	absPath := filepath.Join(repoRoot, filePath)
	if pkg := extractPackage(filePath); pkg != "" {
		sb.WriteString(fmt.Sprintf("  Package: internal/%s\n", pkg))
	}
	sb.WriteString(fmt.Sprintf("  Path: %s\n\n", absPath))

	// Related semantic records by path matching.
	sb.WriteString("RELATED DECISIONS\n")
	decisions, _ := events.ListDecisions(m.storage)
	found := 0
	for _, d := range decisions {
		for _, rf := range d.RelatedFiles {
			if strings.Contains(rf, filePath) || strings.Contains(filePath, rf) {
				sb.WriteString(fmt.Sprintf("  - %s: %s\n", d.Title, d.Decision))
				found++
				break
			}
		}
	}
	if found == 0 {
		sb.WriteString("  None explicitly linked to this file.\n")
	}

	return sb.String(), nil
}

// ExplainWithSession explains changes to a file with session context.
func (m *Manager) ExplainWithSession(filePath, sessionID string) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("FILE: %s\n\n", filePath))

	sessionHistory, err := m.historyManager.GetSessionHistory(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get session history: %w", err)
	}

	sb.WriteString("SESSION CONTEXT\n")
	sb.WriteString(fmt.Sprintf("  Session ID: %s\n", sessionID))
	if sessionHistory.Session.Agent != "" {
		sb.WriteString(fmt.Sprintf("  Agent: %s\n", sessionHistory.Session.Agent))
	}
	if sessionHistory.Session.Model != "" {
		sb.WriteString(fmt.Sprintf("  Model: %s\n", sessionHistory.Session.Model))
	}
	sb.WriteString(fmt.Sprintf("  Started: %s\n\n", sessionHistory.Session.StartTime.Format(time.RFC3339)))

	// Check if file was modified in this session's checkpoints.
	var relevantCheckpoint *checkpoint.Checkpoint
	for i := range sessionHistory.Checkpoints {
		cp := &sessionHistory.Checkpoints[i]
		for _, changedFile := range cp.ChangedFiles {
			if strings.Contains(changedFile, filePath) || strings.Contains(filePath, changedFile) {
				relevantCheckpoint = cp
				break
			}
		}
		if relevantCheckpoint != nil {
			break
		}
	}

	if relevantCheckpoint != nil {
		sb.WriteString("CHECKPOINT CONTEXT\n")
		sb.WriteString(fmt.Sprintf("  Modified in checkpoint: %s\n", relevantCheckpoint.ID))
		sb.WriteString(fmt.Sprintf("  Time: %s\n\n", relevantCheckpoint.Timestamp.Format(time.RFC3339)))

		content, err := m.checkpointManager.GetFileAtCheckpoint(relevantCheckpoint.Ref, filePath)
		if err == nil && content != "" {
			sb.WriteString("FILE CONTENT AT CHECKPOINT:\n")
			sb.WriteString(content)
			sb.WriteString("\n")
		}
	}

	// Related semantic records: filenames in sessionHistory.Decisions/Discoveries
	// are bare filenames (e.g., "abc123-some-title.md"), not full paths.
	sb.WriteString("RELATED SEMANTIC RECORDS\n")

	for _, decisionFile := range sessionHistory.Decisions {
		content, err := m.storage.ReadMarkdown(storage.DecisionsDir + "/" + decisionFile)
		if err != nil {
			continue
		}
		if d := events.ParseDecision(content); d != nil {
			sb.WriteString(fmt.Sprintf("  Decision: %s\n", d.Title))
			if d.Decision != "" {
				sb.WriteString(fmt.Sprintf("    %s\n", d.Decision))
			}
		}
	}

	for _, discoveryFile := range sessionHistory.Discoveries {
		content, err := m.storage.ReadMarkdown(storage.DiscoveriesDir + "/" + discoveryFile)
		if err != nil {
			continue
		}
		if d := events.ParseDiscovery(content); d != nil {
			sb.WriteString(fmt.Sprintf("  Discovery: %s\n", d.Title))
			if d.Finding != "" {
				sb.WriteString(fmt.Sprintf("    %s\n", d.Finding))
			}
		}
	}

	for _, failureFile := range sessionHistory.Failures {
		content, err := m.storage.ReadMarkdown(storage.FailuresDir + "/" + failureFile)
		if err != nil {
			continue
		}
		if f := events.ParseFailure(content); f != nil {
			sb.WriteString(fmt.Sprintf("  Failure: %s\n", f.Title))
			if f.WhyItFailed != "" {
				sb.WriteString(fmt.Sprintf("    %s\n", f.WhyItFailed))
			}
		}
	}

	if sessionHistory.Handoff != "" {
		sb.WriteString("\nSESSION HANDOFF\n")
		sb.WriteString(sessionHistory.Handoff)
	}

	return sb.String(), nil
}
