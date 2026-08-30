package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"agent-ledger/internal/storage"
)

// ExportFormat represents the output format
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
	FormatMarkdown ExportFormat = "markdown"
)

// Exporter handles exporting ledger data
type Exporter struct {
	storage *storage.Storage
}

// NewExporter creates a new exporter
func NewExporter(st *storage.Storage) *Exporter {
	return &Exporter{storage: st}
}

// SessionExportData represents session data for export
type SessionExportData struct {
	ID           string        `json:"id"`
	Agent        string        `json:"agent,omitempty"`
	Model        string        `json:"model,omitempty"`
	Branch       string        `json:"branch"`
	Commit       string        `json:"commit"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      *time.Time    `json:"end_time,omitempty"`
	Decisions    []RecordData  `json:"decisions,omitempty"`
	Discoveries  []RecordData  `json:"discoveries,omitempty"`
	Failures     []RecordData  `json:"failures,omitempty"`
	Constraints  []RecordData  `json:"constraints,omitempty"`
}

// RecordData represents a record for export
type RecordData struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

// ExportOptions contains export configuration
type ExportOptions struct {
	Format        ExportFormat
	IncludeContent bool
	SessionID     string // Optional: export specific session
	RecordType    string // Optional: filter by type (decision, discovery, etc)
}

// ExportSessions exports session data in the specified format
func (e *Exporter) ExportSessions(out io.Writer, opts ExportOptions) error {
	decisions, _ := e.storage.ListFiles("decisions")
	discoveries, _ := e.storage.ListFiles("discoveries")
	failures, _ := e.storage.ListFiles("failures")
	constraints, _ := e.storage.ListFiles("constraints")

	sessionsData := make([]SessionExportData, 0)

	// Collect record data
	decisionRecords := e.loadRecords("decisions", decisions, opts)
	discoveryRecords := e.loadRecords("discoveries", discoveries, opts)
	failureRecords := e.loadRecords("failures", failures, opts)
	constraintRecords := e.loadRecords("constraints", constraints, opts)

	// Create session export
	sessionData := SessionExportData{
		ID:          "project-export",
		StartTime:   time.Now(),
		Decisions:   decisionRecords,
		Discoveries: discoveryRecords,
		Failures:    failureRecords,
		Constraints: constraintRecords,
	}
	sessionsData = append(sessionsData, sessionData)

	switch opts.Format {
	case FormatJSON:
		return e.exportJSON(out, sessionsData)
	case FormatCSV:
		return e.exportCSV(out, sessionsData)
	case FormatMarkdown:
		return e.exportMarkdown(out, sessionsData)
	default:
		return fmt.Errorf("unknown export format: %v", opts.Format)
	}
}

// exportJSON exports data as JSON
func (e *Exporter) exportJSON(out io.Writer, sessions []SessionExportData) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sessions)
}

// exportCSV exports data as CSV
func (e *Exporter) exportCSV(out io.Writer, sessions []SessionExportData) error {
	writer := csv.NewWriter(out)
	defer writer.Flush()

	// Write header
	header := []string{"Type", "Title", "Timestamp", "Path"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write records
	for _, session := range sessions {
		// Write decisions
		for _, record := range session.Decisions {
			if err := writer.Write([]string{
				record.Type,
				record.Title,
				record.Timestamp.Format(time.RFC3339),
				record.Path,
			}); err != nil {
				return err
			}
		}

		// Write discoveries
		for _, record := range session.Discoveries {
			if err := writer.Write([]string{
				record.Type,
				record.Title,
				record.Timestamp.Format(time.RFC3339),
				record.Path,
			}); err != nil {
				return err
			}
		}

		// Write failures
		for _, record := range session.Failures {
			if err := writer.Write([]string{
				record.Type,
				record.Title,
				record.Timestamp.Format(time.RFC3339),
				record.Path,
			}); err != nil {
				return err
			}
		}

		// Write constraints
		for _, record := range session.Constraints {
			if err := writer.Write([]string{
				record.Type,
				record.Title,
				record.Timestamp.Format(time.RFC3339),
				record.Path,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// exportMarkdown exports data as Markdown
func (e *Exporter) exportMarkdown(out io.Writer, sessions []SessionExportData) error {
	fmt.Fprintf(out, "# Agent Ledger Export\n\n")
	fmt.Fprintf(out, "Generated: %s\n\n", time.Now().Format(time.RFC3339))

	for _, session := range sessions {
		fmt.Fprintf(out, "## Session: %s\n\n", session.ID)

		// Decisions
		if len(session.Decisions) > 0 {
			fmt.Fprintf(out, "### Decisions (%d)\n\n", len(session.Decisions))
			for _, d := range session.Decisions {
				fmt.Fprintf(out, "#### %s\n", d.Title)
				fmt.Fprintf(out, "**Date:** %s\n\n", d.Timestamp.Format(time.RFC3339))
				if len(d.Content) > 0 {
					fmt.Fprintf(out, "%s\n\n", d.Content)
				}
			}
		}

		// Discoveries
		if len(session.Discoveries) > 0 {
			fmt.Fprintf(out, "### Discoveries (%d)\n\n", len(session.Discoveries))
			for _, d := range session.Discoveries {
				fmt.Fprintf(out, "#### %s\n", d.Title)
				fmt.Fprintf(out, "**Date:** %s\n\n", d.Timestamp.Format(time.RFC3339))
				if len(d.Content) > 0 {
					fmt.Fprintf(out, "%s\n\n", d.Content)
				}
			}
		}

		// Failures
		if len(session.Failures) > 0 {
			fmt.Fprintf(out, "### Failures (%d)\n\n", len(session.Failures))
			for _, f := range session.Failures {
				fmt.Fprintf(out, "#### %s\n", f.Title)
				fmt.Fprintf(out, "**Date:** %s\n\n", f.Timestamp.Format(time.RFC3339))
				if len(f.Content) > 0 {
					fmt.Fprintf(out, "%s\n\n", f.Content)
				}
			}
		}

		// Constraints
		if len(session.Constraints) > 0 {
			fmt.Fprintf(out, "### Constraints (%d)\n\n", len(session.Constraints))
			for _, c := range session.Constraints {
				fmt.Fprintf(out, "#### %s\n", c.Title)
				fmt.Fprintf(out, "**Date:** %s\n\n", c.Timestamp.Format(time.RFC3339))
				if len(c.Content) > 0 {
					fmt.Fprintf(out, "%s\n\n", c.Content)
				}
			}
		}
	}

	return nil
}

// loadRecords loads records from files
func (e *Exporter) loadRecords(recordType string, files []string, opts ExportOptions) []RecordData {
	records := make([]RecordData, 0)

	if opts.RecordType != "" && opts.RecordType != recordType {
		return records
	}

	for _, file := range files {
		path := recordType + "/" + file
		content, _ := e.storage.ReadMarkdown(path)

		if !opts.IncludeContent {
			content = ""
		}

		title := strings.TrimSuffix(file, ".md")
		records = append(records, RecordData{
			ID:        extractIDFromFilename(file),
			Type:      recordType,
			Title:     title,
			Content:   content,
			Timestamp: time.Now(),
			Path:      path,
		})
	}

	return records
}

// extractIDFromFilename extracts the ID from a filename (assumes format: id-slug.md)
func extractIDFromFilename(filename string) string {
	parts := strings.SplitN(filename, "-", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
