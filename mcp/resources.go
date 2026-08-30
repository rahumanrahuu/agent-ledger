package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"agent-ledger/internal/checkpoint"
	agentcontext "agent-ledger/internal/context"
	"agent-ledger/internal/collaboration"
	"agent-ledger/internal/events"
	"agent-ledger/internal/history"
	"agent-ledger/internal/memory"
	"agent-ledger/internal/quality"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/session"
	"agent-ledger/internal/storage"
)

// Manager manages MCP resources and tools
type Manager struct {
	sessionManager    *session.Manager
	checkpointManager *checkpoint.Manager
	historyManager    *history.Manager
	contextManager    *agentcontext.Manager
	storage           *storage.Storage
	coordinator       *collaboration.Coordinator
	scorer            *quality.Scorer
	memoryFunc        func(string) (*memory.Manager, error)
	repositoryFunc    func() (*repository.Repository, error)
}

// NewManager creates a new MCP manager
func NewManager(st *storage.Storage) (*Manager, error) {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		return nil, err
	}

	sessionManager := session.NewManager(st)
	checkpointManager := checkpoint.NewManager(st)
	historyManager := history.NewManager(sessionManager, checkpointManager, st)
	contextManager := agentcontext.NewManager(historyManager, checkpointManager, st)

	// Create collaboration and quality systems
	coordinator := collaboration.NewCoordinator()
	scorer := quality.NewScorer()

	// Create dependency injection functions
	memoryFunc := func(root string) (*memory.Manager, error) {
		return memory.NewManager(root)
	}
	repositoryFunc := func() (*repository.Repository, error) {
		return repository.Detect()
	}

	return &Manager{
		sessionManager:    sessionManager,
		checkpointManager: checkpointManager,
		historyManager:    historyManager,
		contextManager:    contextManager,
		storage:           st,
		coordinator:       coordinator,
		scorer:            scorer,
		memoryFunc:        memoryFunc,
		repositoryFunc:    repositoryFunc,
	}, nil
}

// RegisterResources registers all MCP resources
func (m *Manager) RegisterResources(server *mcp.Server) error {
	// Register context resource
	server.AddResource(&mcp.Resource{
		URI:         "agent://project/context",
		Name:        "Project Context",
		Description: "Compiled project context including architecture, decisions, discoveries, and recent changes",
		MIMEType:    "text/plain",
	}, m.handleProjectContext)
	
	// Register state resource
	server.AddResource(&mcp.Resource{
		URI:         "agent://project/state",
		Name:        "Project State",
		Description: "Current repository and ledger state",
		MIMEType:    "text/plain",
	}, m.handleProjectState)
	
	// Register architecture resource
	server.AddResource(&mcp.Resource{
		URI:         "agent://project/architecture",
		Name:        "Project Architecture",
		Description: "Project structure and package relationships",
		MIMEType:    "text/plain",
	}, m.handleArchitecture)
	
	// Register decisions resource
	server.AddResource(&mcp.Resource{
		URI:         "agent://project/decisions",
		Name:        "Project Decisions",
		Description: "All architectural and implementation decisions",
		MIMEType:    "text/plain",
	}, m.handleDecisions)
	
	// Register discoveries resource
	server.AddResource(&mcp.Resource{
		URI:         "agent://project/discoveries",
		Name:        "Project Discoveries",
		Description: "All discoveries made during development",
		MIMEType:    "text/plain",
	}, m.handleDiscoveries)
	
	// Register failures resource
	server.AddResource(&mcp.Resource{
		URI:         "agent://project/failures",
		Name:        "Project Failures",
		Description: "All failed approaches and lessons learned",
		MIMEType:    "text/plain",
	}, m.handleFailures)
	
	// Register constraints resource
	server.AddResource(&mcp.Resource{
		URI:         "agent://project/constraints",
		Name:        "Project Constraints",
		Description: "All project constraints and limitations",
		MIMEType:    "text/plain",
	}, m.handleConstraints)
	
	// Register current session resource
	server.AddResource(&mcp.Resource{
		URI:         "agent://session/current",
		Name:        "Current Session",
		Description: "Current active session information",
		MIMEType:    "text/plain",
	}, m.handleCurrentSession)
	
	// Register session history resource
	server.AddResource(&mcp.Resource{
		URI:         "agent://session/history",
		Name:        "Session History",
		Description: "Complete session history",
		MIMEType:    "text/plain",
	}, m.handleSessionHistory)
	
	// Register latest handoff resource
	server.AddResource(&mcp.Resource{
		URI:         "agent://handoff/latest",
		Name:        "Latest Handoff",
		Description: "Most recent session handoff",
		MIMEType:    "text/plain",
	}, m.handleLatestHandoff)
	
	return nil
}

// handleProjectContext handles the project context resource
func (m *Manager) handleProjectContext(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	repo, err := repository.Detect()
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://project/context", Text: fmt.Sprintf("Error detecting repository: %v", err)},
			},
		}, nil
	}
	
	// Use the real context compiler
	compiledCtx, err := m.contextManager.Compile(repo, "")
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://project/context", Text: fmt.Sprintf("Error compiling context: %v", err)},
			},
		}, nil
	}
	
	formatted := m.contextManager.Format(compiledCtx)
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "agent://project/context", Text: formatted},
		},
	}, nil
}

// handleProjectState handles the project state resource
func (m *Manager) handleProjectState(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	repo, err := repository.Detect()
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://project/state", Text: fmt.Sprintf("Error detecting repository: %v", err)},
			},
		}, nil
	}
	
	var sb string
	sb += "PROJECT STATE\n\n"
	sb += fmt.Sprintf("Repository: %s\n", repo.Root)
	sb += fmt.Sprintf("Branch: %s\n", repo.Branch)
	sb += fmt.Sprintf("Head: %s\n", repo.Head)
	sb += fmt.Sprintf("Dirty: %v\n", repo.Dirty)
	
	if repo.Dirty {
		sb += "\nStaged files:\n"
		for _, file := range repo.Staged {
			sb += fmt.Sprintf("  - %s\n", file)
		}
		sb += "\nUnstaged files:\n"
		for _, file := range repo.Unstaged {
			sb += fmt.Sprintf("  - %s\n", file)
		}
		sb += "\nUntracked files:\n"
		for _, file := range repo.Untracked {
			sb += fmt.Sprintf("  - %s\n", file)
		}
	}
	
	// Check current session
	currentSession, err := m.sessionManager.GetCurrent()
	if err == nil {
		sb += "\nActive Session:\n"
		sb += fmt.Sprintf("  ID: %s\n", currentSession.ID)
		sb += fmt.Sprintf("  Agent: %s\n", currentSession.Agent)
		if currentSession.Model != "" {
			sb += fmt.Sprintf("  Model: %s\n", currentSession.Model)
		}
		sb += fmt.Sprintf("  Started: %s\n", currentSession.StartTime.Format("2006-01-02 15:04:05"))
	} else {
		sb += "\nNo active session\n"
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "agent://project/state", Text: sb},
		},
	}, nil
}

// handleArchitecture handles the architecture resource
func (m *Manager) handleArchitecture(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	repo, err := repository.Detect()
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://project/architecture", Text: fmt.Sprintf("Error detecting repository: %v", err)},
			},
		}, nil
	}
	
	compiledCtx, err := m.contextManager.Compile(repo, "")
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://project/architecture", Text: fmt.Sprintf("Error compiling context: %v", err)},
			},
		}, nil
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "agent://project/architecture", Text: compiledCtx.Architecture},
		},
	}, nil
}

// handleDecisions handles the decisions resource
func (m *Manager) handleDecisions(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	allDecisions, err := events.ListDecisions(m.storage)
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://project/decisions", Text: fmt.Sprintf("Error listing decisions: %v", err)},
			},
		}, nil
	}
	
	var sb string
	if len(allDecisions) == 0 {
		sb = "No decisions recorded yet."
	} else {
		sb = fmt.Sprintf("DECISIONS (%d)\n\n", len(allDecisions))
		for _, d := range allDecisions {
			sb += fmt.Sprintf("## %s\n\n**Decision:** %s\n**Rationale:** %s\n\n", d.Title, d.Decision, d.Rationale)
		}
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "agent://project/decisions", Text: sb},
		},
	}, nil
}

// handleDiscoveries handles the discoveries resource
func (m *Manager) handleDiscoveries(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	allDiscoveries, err := events.ListDiscoveries(m.storage)
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://project/discoveries", Text: fmt.Sprintf("Error listing discoveries: %v", err)},
			},
		}, nil
	}
	
	var sb string
	if len(allDiscoveries) == 0 {
		sb = "No discoveries recorded yet."
	} else {
		sb = fmt.Sprintf("DISCOVERIES (%d)\n\n", len(allDiscoveries))
		for _, d := range allDiscoveries {
			sb += fmt.Sprintf("## %s\n\n**Finding:** %s\n\n", d.Title, d.Finding)
		}
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "agent://project/discoveries", Text: sb},
		},
	}, nil
}

// handleFailures handles the failures resource
func (m *Manager) handleFailures(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	allFailures, err := events.ListFailures(m.storage)
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://project/failures", Text: fmt.Sprintf("Error listing failures: %v", err)},
			},
		}, nil
	}
	
	var sb string
	if len(allFailures) == 0 {
		sb = "No failures recorded yet."
	} else {
		sb = fmt.Sprintf("FAILURES (%d)\n\n", len(allFailures))
		for _, f := range allFailures {
			sb += fmt.Sprintf("## %s\n\n**Attempted Approach:** %s\n**Why It Failed:** %s\n", f.Title, f.AttemptedApproach, f.WhyItFailed)
			if f.Lessons != "" {
				sb += fmt.Sprintf("**Lessons:** %s\n", f.Lessons)
			}
			sb += "\n"
		}
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "agent://project/failures", Text: sb},
		},
	}, nil
}

// handleConstraints handles the constraints resource
func (m *Manager) handleConstraints(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	allConstraints, err := events.ListConstraints(m.storage)
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://project/constraints", Text: fmt.Sprintf("Error listing constraints: %v", err)},
			},
		}, nil
	}
	
	var sb string
	if len(allConstraints) == 0 {
		sb = "No constraints recorded yet."
	} else {
		sb = fmt.Sprintf("CONSTRAINTS (%d)\n\n", len(allConstraints))
		for _, c := range allConstraints {
			sb += fmt.Sprintf("## %s\n\n**Constraint:** %s\n**Reason:** %s\n\n", c.Title, c.Constraint, c.Reason)
		}
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "agent://project/constraints", Text: sb},
		},
	}, nil
}

// handleCurrentSession handles the current session resource
func (m *Manager) handleCurrentSession(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	session, err := m.sessionManager.GetCurrent()
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://session/current", Text: "No active session"},
			},
		}, nil
	}
	
	content := fmt.Sprintf("Session ID: %s\nAgent: %s\nModel: %s\nStarted: %s", 
		session.ID, session.Agent, session.Model, session.StartTime.Format("2006-01-02 15:04:05"))
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "agent://session/current", Text: content},
		},
	}, nil
}

// handleSessionHistory handles the session history resource
func (m *Manager) handleSessionHistory(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	sessions, err := m.historyManager.GetAllSessions("", "")
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://session/history", Text: fmt.Sprintf("Error getting sessions: %v", err)},
			},
		}, nil
	}
	
	var sb string
	if len(sessions) == 0 {
		sb = "No sessions found"
	} else {
		sb = "SESSION HISTORY\n\n"
		for _, sess := range sessions {
			sb += fmt.Sprintf("Session ID: %s\n", sess.ID)
			if sess.Agent != "" {
				sb += fmt.Sprintf("  Agent: %s\n", sess.Agent)
			}
			if sess.Model != "" {
				sb += fmt.Sprintf("  Model: %s\n", sess.Model)
			}
			sb += fmt.Sprintf("  Branch: %s\n", sess.Branch)
			sb += fmt.Sprintf("  Started: %s\n", sess.StartTime.Format("2006-01-02 15:04:05"))
			if sess.EndTime != nil {
				sb += fmt.Sprintf("  Ended: %s\n", sess.EndTime.Format("2006-01-02 15:04:05"))
			} else {
				sb += "  Status: Active\n"
			}
			sb += "\n"
		}
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "agent://session/history", Text: sb},
		},
	}, nil
}

// handleLatestHandoff handles the latest handoff resource
func (m *Manager) handleLatestHandoff(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	repo, err := repository.Detect()
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://handoff/latest", Text: fmt.Sprintf("Error detecting repository: %v", err)},
			},
		}, nil
	}
	
	compiledCtx, err := m.contextManager.Compile(repo, "")
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://handoff/latest", Text: fmt.Sprintf("Error compiling context: %v", err)},
			},
		}, nil
	}
	
	if compiledCtx.LatestHandoff == "" {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "agent://handoff/latest", Text: "No handoff found"},
			},
		}, nil
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "agent://handoff/latest", Text: compiledCtx.LatestHandoff},
		},
	}, nil
}
