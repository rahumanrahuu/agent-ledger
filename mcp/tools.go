package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"agent-ledger/internal/events"
	"agent-ledger/internal/git"
	"agent-ledger/internal/memory"
	"agent-ledger/internal/repository"
)

// MCP tool input schemas.
//
// These are raw JSON schemas because RegisterTools uses the low-level
// mcp.Server.AddTool API. The SDK requires a non-nil object schema for
// every tool, including tools that take no arguments.
var (
	emptyInputSchema = json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`)

	startSessionSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"agent": {
				"type": "string",
				"description": "Optional AI coding agent or IDE name (e.g. 'Antigravity IDE', 'Claude Code', 'Cursor AI', 'Codex')"
			},
			"model": {
				"type": "string",
				"description": "Optional AI model identifier (e.g. 'gemini-3.6-flash', 'claude-3-7-sonnet', 'gpt-4o')"
			}
		}
	}`)

	contextInputSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"task": {
				"type": "string",
				"description": "Optional task description used to prioritize relevant context"
			}
		}
	}`)

	historyInputSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"session_id": {
				"type": "string",
				"description": "Optional session ID to retrieve a specific session"
			}
		}
	}`)

	explainFileSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"file_path": {
				"type": "string",
				"description": "Path to the file to explain"
			}
		},
		"required": ["file_path"]
	}`)

	recordDecisionSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"title": {
				"type": "string",
				"description": "Title of the decision"
			},
			"decision": {
				"type": "string",
				"description": "The decision that was made"
			},
			"rationale": {
				"type": "string",
				"description": "Rationale for the decision"
			}
		},
		"required": ["title", "decision"]
	}`)

	recordDiscoverySchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"title": {
				"type": "string",
				"description": "Title of the discovery"
			},
			"finding": {
				"type": "string",
				"description": "What was discovered"
			}
		},
		"required": ["title", "finding"]
	}`)

	recordFailureSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"title": {
				"type": "string",
				"description": "Title of the failure"
			},
			"attempted": {
				"type": "string",
				"description": "What was attempted"
			},
			"why": {
				"type": "string",
				"description": "Why the attempt failed"
			},
			"lessons": {
				"type": "string",
				"description": "Lessons learned from the failure"
			}
		},
		"required": ["title", "attempted", "why"]
	}`)

	recordConstraintSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"title": {
				"type": "string",
				"description": "Title of the constraint"
			},
			"constraint": {
				"type": "string",
				"description": "The constraint future work must preserve"
			},
			"reason": {
				"type": "string",
				"description": "Reason for the constraint"
			}
		},
		"required": ["title", "constraint"]
	}`)

	createHandoffSchema = json.RawMessage(`{
		"type": "object",
		"properties": {
			"current_state": {
				"type": "string",
				"description": "Current state of the project"
			},
			"what_changed": {
				"type": "string",
				"description": "What changed during this session"
			}
		},
		"required": ["current_state", "what_changed"]
	}`)
	createMemorySchema = json.RawMessage(`{"type":"object","properties":{"type":{"type":"string"},"title":{"type":"string"},"content":{"type":"string"},"keywords":{"type":"string"},"importance":{"type":"number"}},"required":["type","title","content"]}`)
)

// RegisterTools registers all Agent Ledger MCP tools.
func (m *Manager) RegisterTools(server *mcp.Server) error {
	server.AddTool(&mcp.Tool{
		Name:        "start_session",
		Description: "Start an Agent Ledger session with optional agent and model metadata",
		InputSchema: startSessionSchema,
	}, m.handleStartSession)

	server.AddTool(&mcp.Tool{
		Name:        "checkpoint",
		Description: "Create a Git-native checkpoint of the current Agent Ledger session",
		InputSchema: emptyInputSchema,
	}, m.handleCheckpoint)

	server.AddTool(&mcp.Tool{
		Name:        "get_context",
		Description: "Get compiled project context, optionally prioritized for a specific task",
		InputSchema: contextInputSchema,
	}, m.handleGetContext)

	server.AddTool(&mcp.Tool{
		Name:        "get_history",
		Description: "Get Agent Ledger session history, optionally for a specific session",
		InputSchema: historyInputSchema,
	}, m.handleGetHistory)

	server.AddTool(&mcp.Tool{
		Name:        "get_handoff",
		Description: "Get the latest Agent Ledger handoff",
		InputSchema: emptyInputSchema,
	}, m.handleGetHandoff)

	server.AddTool(&mcp.Tool{
		Name:        "explain_file",
		Description: "Explain the development history and context of a specific file",
		InputSchema: explainFileSchema,
	}, m.handleExplainFile)

	server.AddTool(&mcp.Tool{
		Name:        "record_decision",
		Description: "Record an important architectural or implementation decision",
		InputSchema: recordDecisionSchema,
	}, m.handleRecordDecision)

	server.AddTool(&mcp.Tool{
		Name:        "record_discovery",
		Description: "Record an important discovery about the codebase or development process",
		InputSchema: recordDiscoverySchema,
	}, m.handleRecordDiscovery)

	server.AddTool(&mcp.Tool{
		Name:        "record_failure",
		Description: "Record a failed approach and what was learned from it",
		InputSchema: recordFailureSchema,
	}, m.handleRecordFailure)

	server.AddTool(&mcp.Tool{
		Name:        "record_constraint",
		Description: "Record a constraint that future agents should preserve",
		InputSchema: recordConstraintSchema,
	}, m.handleRecordConstraint)

	server.AddTool(&mcp.Tool{
		Name:        "create_handoff",
		Description: "Create a handoff for the next development session",
		InputSchema: createHandoffSchema,
	}, m.handleCreateHandoff)
	server.AddTool(&mcp.Tool{Name: "create_memory", Description: "Store durable project knowledge", InputSchema: createMemorySchema}, m.handleCreateMemory)

	server.AddTool(&mcp.Tool{
		Name:        "validate",
		Description: "Validate Agent Ledger integrity and consistency",
		InputSchema: emptyInputSchema,
	}, m.handleValidate)

	return nil
}

func (m *Manager) handleCreateMemory(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	typeName, title, content := m.getStringParam(req, "type"), m.getStringParam(req, "title"), m.getStringParam(req, "content")
	if typeName == "" || title == "" || content == "" { return m.errorResult("Missing required parameters: type, title, and content"), nil }
	importance := 0.5; var args map[string]json.RawMessage
	if req != nil && req.Params != nil && json.Unmarshal(req.Params.Arguments, &args) == nil { _ = json.Unmarshal(args["importance"], &importance) }
	repo, err := repository.Detect(); if err != nil { return m.errorResult(err.Error()), nil }
	mgr, err := memory.NewManager(repo.Root); if err != nil { return m.errorResult(err.Error()), nil }; defer mgr.Close()
	sessionID := ""; if current, err := m.sessionManager.GetCurrent(); err == nil { sessionID = current.ID }
	id := uuid.New().String(); err = mgr.Add(memory.Memory{ID:id, Type:typeName, Title:title, Content:content, Keywords:m.getStringParam(req,"keywords"), Importance:importance, SessionID:sessionID, Path:"mcp://create_memory"})
	if err != nil { return m.errorResult(err.Error()), nil }; return m.textResult(fmt.Sprintf("Memory created successfully\nID: %s", id)), nil
}

// handleStartSession starts a new Agent Ledger session.
func (m *Manager) handleStartSession(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	agent := m.getStringParam(req, "agent")
	model := m.getStringParam(req, "model")

	repo, err := repository.Detect()
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error detecting repository: %v", err)), nil
	}

	newSession, err := m.sessionManager.Create(
		agent,
		model,
		repo.Root,
		repo.Branch,
		repo.Head,
	)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error creating session: %v", err)), nil
	}

	return m.textResult(fmt.Sprintf(
		"Session started successfully\nID: %s\nAgent: %s\nModel: %s\nBranch: %s\nHead: %s",
		newSession.ID,
		valueOrUnspecified(agent),
		valueOrUnspecified(model),
		repo.Branch,
		repo.Head,
	)), nil
}

// handleCheckpoint creates a checkpoint.
func (m *Manager) handleCheckpoint(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	repo, err := repository.Detect()
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error detecting repository: %v", err)), nil
	}

	currentSession, err := m.sessionManager.GetCurrent()
	if err != nil {
		return m.errorResult("No active session. Start a session first."), nil
	}

	cp, err := m.checkpointManager.Create(currentSession.ID, repo)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error creating checkpoint: %v", err)), nil
	}

	return m.textResult(fmt.Sprintf(
		"Checkpoint created successfully\nID: %s\nSession: %s\nCommit: %s\nRef: %s\nChanged files: %d",
		cp.ID,
		cp.SessionID,
		cp.Commit,
		cp.Ref,
		len(cp.ChangedFiles),
	)), nil
}

// handleGetContext gets compiled project context.
func (m *Manager) handleGetContext(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	repo, err := repository.Detect()
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error detecting repository: %v", err)), nil
	}

	task := m.getStringParam(req, "task")

	compiledCtx, err := m.contextManager.Compile(repo, task)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error compiling context: %v", err)), nil
	}

	formatted := m.contextManager.Format(compiledCtx)

	return m.textResult(formatted), nil
}

// handleGetHistory gets session/project history.
func (m *Manager) handleGetHistory(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	sessionID := m.getStringParam(req, "session_id")

	sessions, err := m.historyManager.GetAllSessions("", "")
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error getting sessions: %v", err)), nil
	}

	if sessionID != "" {
		filtered := sessions[:0]
		for _, session := range sessions {
			if session.ID == sessionID {
				filtered = append(filtered, session)
				break
			}
		}
		sessions = filtered
	}

	if len(sessions) == 0 {
		return m.textResult("No sessions found"), nil
	}

	var sb strings.Builder
	sb.WriteString("SESSION HISTORY\n\n")

	for _, sess := range sessions {
		fmt.Fprintf(&sb, "Session ID: %s\n", sess.ID)

		if sess.Agent != "" {
			fmt.Fprintf(&sb, "  Agent: %s\n", sess.Agent)
		}

		if sess.Model != "" {
			fmt.Fprintf(&sb, "  Model: %s\n", sess.Model)
		}

		fmt.Fprintf(&sb, "  Branch: %s\n", sess.Branch)
		fmt.Fprintf(&sb, "  Started: %s\n", sess.StartTime.Format("2006-01-02 15:04:05"))

		if sess.EndTime != nil {
			fmt.Fprintf(&sb, "  Ended: %s\n", sess.EndTime.Format("2006-01-02 15:04:05"))
		} else {
			sb.WriteString("  Status: Active\n")
		}

		sb.WriteString("\n")
	}

	return m.textResult(sb.String()), nil
}

// handleGetHandoff gets the latest handoff.
func (m *Manager) handleGetHandoff(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	repo, err := repository.Detect()
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error detecting repository: %v", err)), nil
	}

	compiledCtx, err := m.contextManager.Compile(repo, "")
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error compiling context: %v", err)), nil
	}

	if strings.TrimSpace(compiledCtx.LatestHandoff) == "" {
		return m.textResult("No handoff found"), nil
	}

	return m.textResult(compiledCtx.LatestHandoff), nil
}

// handleExplainFile explains the development history of a file.
func (m *Manager) handleExplainFile(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	filePath := m.getStringParam(req, "file_path")

	if filePath == "" {
		return m.errorResult("Missing required parameter: file_path"), nil
	}

	currentSession, err := m.sessionManager.GetCurrent()
	if err == nil {
		explanation, err := m.contextManager.ExplainWithSession(
			filePath,
			currentSession.ID,
		)
		if err != nil {
			return m.errorResult(fmt.Sprintf("Error explaining file: %v", err)), nil
		}

		return m.textResult(explanation), nil
	}

	explanation, err := m.contextManager.Explain(filePath)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error explaining file: %v", err)), nil
	}

	return m.textResult(explanation), nil
}

// handleRecordDecision records a decision.
func (m *Manager) handleRecordDecision(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	currentSession, err := m.sessionManager.GetCurrent()
	if err != nil {
		return m.errorResult("No active session. Start a session first."), nil
	}

	title := m.getStringParam(req, "title")
	decision := m.getStringParam(req, "decision")
	rationale := m.getStringParam(req, "rationale")

	if title == "" || decision == "" {
		return m.errorResult("Missing required parameters: title and decision"), nil
	}

	eventsManager := events.NewManager(m.storage)

	decisionRecord, err := eventsManager.CreateDecision(
		currentSession.ID,
		title,
		decision,
		rationale,
		[]string{},
		[]string{},
	)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error creating decision: %v", err)), nil
	}

	repo, err := repository.Detect()
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error detecting repository: %v", err)), nil
	}

	mgr, err := memory.NewManager(repo.Root)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error initializing memory manager: %v", err)), nil
	}
	defer mgr.Close()

	content := decision
	if rationale != "" {
		content = decision + `

Rationale: ` + rationale
	}

	err = mgr.Add(memory.Memory{
		ID:         decisionRecord.ID,
		Type:       "decision",
		Title:      title,
		Content:    content,
		SessionID:  currentSession.ID,
		Path:       "mcp://record_decision",
		Importance: 0.7,
	})
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error saving decision to memory: %v", err)), nil
	}

	return m.textResult("Decision recorded successfully"), nil
}

// handleRecordDiscovery records a discovery.
func (m *Manager) handleRecordDiscovery(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	currentSession, err := m.sessionManager.GetCurrent()
	if err != nil {
		return m.errorResult("No active session. Start a session first."), nil
	}

	title := m.getStringParam(req, "title")
	finding := m.getStringParam(req, "finding")

	if title == "" || finding == "" {
		return m.errorResult("Missing required parameters: title and finding"), nil
	}

	eventsManager := events.NewManager(m.storage)

	discoveryRecord, err := eventsManager.CreateDiscovery(
		currentSession.ID,
		title,
		finding,
		[]string{},
		[]string{},
	)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error creating discovery: %v", err)), nil
	}

	repo, err := repository.Detect()
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error detecting repository: %v", err)), nil
	}

	mgr, err := memory.NewManager(repo.Root)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error initializing memory manager: %v", err)), nil
	}
	defer mgr.Close()

	err = mgr.Add(memory.Memory{
		ID:         discoveryRecord.ID,
		Type:       "discovery",
		Title:      title,
		Content:    finding,
		SessionID:  currentSession.ID,
		Path:       "mcp://record_discovery",
		Importance: 0.7,
	})
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error saving discovery to memory: %v", err)), nil
	}

	return m.textResult("Discovery recorded successfully"), nil
}

// handleRecordFailure records a failure.
func (m *Manager) handleRecordFailure(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	currentSession, err := m.sessionManager.GetCurrent()
	if err != nil {
		return m.errorResult("No active session. Start a session first."), nil
	}

	title := m.getStringParam(req, "title")
	attempted := m.getStringParam(req, "attempted")
	why := m.getStringParam(req, "why")
	lessons := m.getStringParam(req, "lessons")

	if title == "" || attempted == "" || why == "" {
		return m.errorResult(
			"Missing required parameters: title, attempted, and why",
		), nil
	}

	eventsManager := events.NewManager(m.storage)

	failureRecord, err := eventsManager.CreateFailure(
		currentSession.ID,
		title,
		attempted,
		why,
		lessons,
	)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error creating failure: %v", err)), nil
	}

	repo, err := repository.Detect()
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error detecting repository: %v", err)), nil
	}

	mgr, err := memory.NewManager(repo.Root)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error initializing memory manager: %v", err)), nil
	}
	defer mgr.Close()

	content := "Attempted: " + attempted + `

Why it failed: ` + why
	if lessons != "" {
		content = content + `

Lessons: ` + lessons
	}

	err = mgr.Add(memory.Memory{
		ID:         failureRecord.ID,
		Type:       "failure",
		Title:      title,
		Content:    content,
		SessionID:  currentSession.ID,
		Path:       "mcp://record_failure",
		Importance: 0.6,
	})
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error saving failure to memory: %v", err)), nil
	}

	return m.textResult("Failure recorded successfully"), nil
}

// handleRecordConstraint records a constraint.
func (m *Manager) handleRecordConstraint(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	currentSession, err := m.sessionManager.GetCurrent()
	if err != nil {
		return m.errorResult("No active session. Start a session first."), nil
	}

	title := m.getStringParam(req, "title")
	constraint := m.getStringParam(req, "constraint")
	reason := m.getStringParam(req, "reason")

	if title == "" || constraint == "" {
		return m.errorResult(
			"Missing required parameters: title and constraint",
		), nil
	}

	eventsManager := events.NewManager(m.storage)

	constraintRecord, err := eventsManager.CreateConstraint(
		currentSession.ID,
		title,
		constraint,
		reason,
	)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error creating constraint: %v", err)), nil
	}

	repo, err := repository.Detect()
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error detecting repository: %v", err)), nil
	}

	mgr, err := memory.NewManager(repo.Root)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error initializing memory manager: %v", err)), nil
	}
	defer mgr.Close()

	content := constraint
	if reason != "" {
		content = constraint + `

Reason: ` + reason
	}

	err = mgr.Add(memory.Memory{
		ID:         constraintRecord.ID,
		Type:       "constraint",
		Title:      title,
		Content:    content,
		SessionID:  currentSession.ID,
		Path:       "mcp://record_constraint",
		Importance: 0.8,
	})
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error saving constraint to memory: %v", err)), nil
	}

	return m.textResult("Constraint recorded successfully"), nil
}

// handleCreateHandoff creates a handoff.
func (m *Manager) handleCreateHandoff(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	currentSession, err := m.sessionManager.GetCurrent()
	if err != nil {
		return m.errorResult("No active session. Start a session first."), nil
	}

	currentState := m.getStringParam(req, "current_state")
	whatChanged := m.getStringParam(req, "what_changed")

	if currentState == "" || whatChanged == "" {
		return m.errorResult(
			"Missing required parameters: current_state and what_changed",
		), nil
	}

	eventsManager := events.NewManager(m.storage)

	_, err = eventsManager.CreateHandoff(
		currentSession.ID,
		currentState,
		whatChanged,
		[]string{},
		[]string{},
		[]string{},
		[]string{},
		"",
		"",
		[]string{},
		[]string{},
	)
	if err != nil {
		return m.errorResult(fmt.Sprintf("Error creating handoff: %v", err)), nil
	}

	return m.textResult("Handoff created successfully"), nil
}

// handleValidate validates ledger integrity.
func (m *Manager) handleValidate(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	var issues []string

	// Validate session records.
	sessions, err := m.storage.ListDirectories("sessions")
	if err != nil {
		issues = append(
			issues,
			fmt.Sprintf("Failed to list sessions: %v", err),
		)
	} else {
		for _, sessionID := range sessions {
			metadataPath := fmt.Sprintf(
				"sessions/%s/metadata.json",
				sessionID,
			)

			if !m.storage.FileExists(metadataPath) {
				issues = append(
					issues,
					fmt.Sprintf(
						"Session %s missing metadata.json",
						sessionID,
					),
				)
			}
		}
	}

	// Validate checkpoint refs.
	refOutput, err := git.Command("show-ref")
	if err == nil {
		refs := strings.Split(refOutput, "\n")
		for _, ref := range refs {
			if strings.Contains(ref, " refs/agents/sessions/") {
				parts := strings.Fields(ref)
				if len(parts) < 2 {
					continue
				}

				if _, err := git.Command("cat-file", "-e", parts[0]); err != nil {
					issues = append(
						issues,
						fmt.Sprintf(
							"Broken checkpoint ref: %s",
							parts[1],
						),
					)
				}
			}
		}
	}

	// Validate semantic record collections.
	recordCollections := []struct {
		name string
	}{
		{name: "decisions"},
		{name: "discoveries"},
		{name: "failures"},
		{name: "constraints"},
	}

	for _, collection := range recordCollections {
		files, err := m.storage.ListFiles(collection.name)
		if err != nil {
			issues = append(
				issues,
				fmt.Sprintf(
					"Failed to list %s: %v",
					collection.name,
					err,
				),
			)
			continue
		}

		for _, file := range files {
			path := collection.name + "/" + file

			content, err := m.storage.ReadMarkdown(path)
			if err != nil {
				issues = append(
					issues,
					fmt.Sprintf(
						"Failed to read %s: %v",
						path,
						err,
					),
				)
				continue
			}

			if strings.TrimSpace(content) == "" {
				issues = append(
					issues,
					fmt.Sprintf(
						"%s is empty",
						path,
					),
				)
			}
		}
	}

	if len(issues) == 0 {
		return m.textResult("Validation passed: No issues found"), nil
	}

	var sb strings.Builder
	fmt.Fprintf(
		&sb,
		"Validation found %d issue(s):\n",
		len(issues),
	)

	for _, issue := range issues {
		fmt.Fprintf(&sb, "  - %s\n", issue)
	}

	return m.textResult(sb.String()), nil
}

// getStringParam extracts a string parameter from the raw MCP arguments.
//
// The official Go MCP SDK's low-level Server.AddTool API passes tool
// arguments to handlers as json.RawMessage, so the handler is responsible
// for unmarshaling them.
func (m *Manager) getStringParam(
	req *mcp.CallToolRequest,
	key string,
) string {
	if req == nil || req.Params == nil || len(req.Params.Arguments) == 0 {
		return ""
	}

	var args map[string]json.RawMessage

	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return ""
	}

	raw, ok := args[key]
	if !ok {
		return ""
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}

	return value
}

// textResult creates a successful text-only tool result.
func (m *Manager) textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: text,
			},
		},
	}
}

// errorResult creates a tool-error result.
func (m *Manager) errorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: message,
			},
		},
		IsError: true,
	}
}

// valueOrUnspecified avoids producing misleading blank metadata in output.
func valueOrUnspecified(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unspecified"
	}

	return value
}
