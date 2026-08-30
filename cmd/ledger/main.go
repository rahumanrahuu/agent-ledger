package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"agent-ledger/internal/checkpoint"
	agentcontext "agent-ledger/internal/context"
	"agent-ledger/internal/events"
	"agent-ledger/internal/git"
	"agent-ledger/internal/history"
	mcpinternal "agent-ledger/internal/mcp"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/session"
	"agent-ledger/internal/storage"
	"agent-ledger/internal/validator"
	"agent-ledger/ui"
)

var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit()
	case "status":
		handleStatus()
	case "start":
		handleStart()
	case "stop":
		handleStop()
	case "checkpoint":
		handleCheckpoint()
	case "history":
		handleHistory()
	case "sessions":
		handleSessions()
	case "context":
		handleContext()
	case "decision":
		handleDecision()
	case "discovery":
		handleDiscovery()
	case "failure":
		handleFailure()
	case "constraint":
		handleConstraint()
	case "handoff":
		handleHandoff()
	case "explain":
		handleExplain()
	case "validate":
		handleValidate()
	case "ui":
		handleUI()
	case "mcp":
		handleMCP()
	case "--help", "-h", "help":
		printUsage()
		os.Exit(0)
	case "--version", "-v", "version":
		fmt.Printf("agent-ledger %s\n", Version)
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Agent Ledger - Git-native AI agent session management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  agent-ledger init              Initialize agent ledger in repository")
	fmt.Println("  agent-ledger status            Show repository and ledger status")
	fmt.Println("  agent-ledger start [--agent <name>] [--model <name>]  Start a new session")
	fmt.Println("  agent-ledger stop              Stop current session")
	fmt.Println("  agent-ledger checkpoint        Create a checkpoint")
	fmt.Println("  agent-ledger history           Show session history")
	fmt.Println("  agent-ledger sessions          List all sessions")
	fmt.Println("  agent-ledger context [--task <task>]  Compile project context")
	fmt.Println("  agent-ledger decision          Record a decision")
	fmt.Println("  agent-ledger discovery         Record a discovery")
	fmt.Println("  agent-ledger failure           Record a failure")
	fmt.Println("  agent-ledger constraint        Record a constraint")
	fmt.Println("  agent-ledger handoff           Create a handoff")
	fmt.Println("  agent-ledger explain <file>      Explain development history of a file")
	fmt.Println("  agent-ledger validate            Validate ledger integrity")
	fmt.Println("  agent-ledger ui [--port <port>] [--host <host>] [--open]  Launch local web UI")
	fmt.Println("  agent-ledger mcp               Start MCP server over stdio")
	fmt.Println("  agent-ledger --help, -h        Show this help message")
	fmt.Println("  agent-ledger --version, -v     Show version")
}

func handleInit() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Initialize storage
	storage := storage.New(repo.Root)
	if storage.Exists() {
		fmt.Println("Agent ledger already initialized")
		return
	}

	if err := storage.Initialize(); err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	// Create project.md
	projectContent := fmt.Sprintf("# Project Information\n\n"+
		"**Repository:** %s\n"+
		"**Branch:** %s\n"+
		"**Head:** %s\n\n"+
		"Initialized: %s\n",
		repo.Root, repo.Branch, repo.Head, getCurrentTime())

	if err := storage.WriteMarkdown("project.md", projectContent); err != nil {
		fmt.Printf("Error writing project.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Agent ledger initialized successfully")
	fmt.Printf("Storage directory: %s\n", storage.GetRoot())
}

func handleStatus() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Repository Status:")
	fmt.Printf("  Root: %s\n", repo.Root)
	fmt.Printf("  Branch: %s\n", repo.Branch)
	fmt.Printf("  Head: %s\n", repo.Head)

	if len(repo.Remotes) > 0 {
		fmt.Println("  Remotes:")
		for name, url := range repo.Remotes {
			fmt.Printf("    %s: %s\n", name, url)
		}
	}

	fmt.Printf("  Dirty: %v\n", repo.Dirty)

	if repo.Dirty {
		if len(repo.Staged) > 0 {
			fmt.Println("  Staged files:")
			for _, file := range repo.Staged {
				fmt.Printf("    - %s\n", file)
			}
		}
		if len(repo.Unstaged) > 0 {
			fmt.Println("  Unstaged files:")
			for _, file := range repo.Unstaged {
				fmt.Printf("    - %s\n", file)
			}
		}
		if len(repo.Untracked) > 0 {
			fmt.Println("  Untracked files:")
			for _, file := range repo.Untracked {
				fmt.Printf("    - %s\n", file)
			}
		}
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("\nAgent Ledger: Not initialized")
		return
	}

	fmt.Println("\nAgent Ledger: Initialized")
	fmt.Printf("  Storage: %s\n", storage.GetRoot())

	// Check for active session
	sessionManager := session.NewManager(storage)
	currentSession, err := sessionManager.GetCurrent()
	if err == nil {
		fmt.Printf("  Active session: %s\n", currentSession.ID)
		if currentSession.Agent != "" {
			fmt.Printf("    Agent: %s\n", currentSession.Agent)
		}
		if currentSession.Model != "" {
			fmt.Printf("    Model: %s\n", currentSession.Model)
		}
		fmt.Printf("    Started: %s\n", currentSession.StartTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("  Active session: None")
	}
}

func handleStart() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Parse flags
	var agent, model string
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--agent":
			if i+1 < len(args) {
				agent = args[i+1]
				i++
			}
		case "--model":
			if i+1 < len(args) {
				model = args[i+1]
				i++
			}
		}
	}

	// Create session
	sessionManager := session.NewManager(storage)
	newSession, err := sessionManager.Create(agent, model, repo.Root, repo.Branch, repo.Head)
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Session started successfully")
	fmt.Printf("Session ID: %s\n", newSession.ID)
	if agent != "" {
		fmt.Printf("Agent: %s\n", agent)
	}
	if model != "" {
		fmt.Printf("Model: %s\n", model)
	}
	fmt.Printf("Branch: %s\n", repo.Branch)
	fmt.Printf("Head: %s\n", repo.Head)
}

func handleStop() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Stop session
	sessionManager := session.NewManager(storage)
	if err := sessionManager.Stop(); err != nil {
		fmt.Printf("Error stopping session: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Session stopped successfully")
}

func handleCheckpoint() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Get current session
	sessionManager := session.NewManager(storage)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Println("No active session. Run 'agent-ledger start' first.")
		os.Exit(1)
	}

	// Create checkpoint
	checkpointManager := checkpoint.NewManager(storage)
	checkpoint, err := checkpointManager.Create(currentSession.ID, repo)
	if err != nil {
		fmt.Printf("Error creating checkpoint: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Checkpoint created successfully")
	fmt.Printf("Checkpoint ID: %s\n", checkpoint.ID)
	fmt.Printf("Session: %s\n", checkpoint.SessionID)
	fmt.Printf("Commit: %s\n", checkpoint.Commit)
	fmt.Printf("Ref: %s\n", checkpoint.Ref)
	fmt.Printf("Changed files: %d\n", len(checkpoint.ChangedFiles))
}

func handleHistory() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Create managers
	sessionManager := session.NewManager(storage)
	checkpointManager := checkpoint.NewManager(storage)
	historyManager := history.NewManager(sessionManager, checkpointManager, storage)

	// Get all sessions
	sessions, err := historyManager.GetAllSessions("", "")
	if err != nil {
		fmt.Printf("Error getting sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return
	}

	fmt.Println("Session History:")
	fmt.Println()

	for _, sess := range sessions {
		fmt.Printf("Session ID: %s\n", sess.ID)
		if sess.Agent != "" {
			fmt.Printf("  Agent: %s\n", sess.Agent)
		}
		if sess.Model != "" {
			fmt.Printf("  Model: %s\n", sess.Model)
		}
		fmt.Printf("  Branch: %s\n", sess.Branch)
		fmt.Printf("  Head: %s\n", sess.Head)
		fmt.Printf("  Started: %s\n", sess.StartTime.Format("2006-01-02 15:04:05"))
		if sess.EndTime != nil {
			fmt.Printf("  Ended: %s\n", sess.EndTime.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("  Status: Active\n")
		}

		// Get session details
		sessionHistory, err := historyManager.GetSessionHistory(sess.ID)
		if err == nil {
			fmt.Printf("  Checkpoints: %d\n", len(sessionHistory.Checkpoints))
			fmt.Printf("  Decisions: %d\n", len(sessionHistory.Decisions))
			fmt.Printf("  Discoveries: %d\n", len(sessionHistory.Discoveries))
			fmt.Printf("  Failures: %d\n", len(sessionHistory.Failures))
			fmt.Printf("  Constraints: %d\n", len(sessionHistory.Constraints))
		}

		fmt.Println()
	}
}

func handleSessions() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Create managers
	sessionManager := session.NewManager(storage)
	checkpointManager := checkpoint.NewManager(storage)
	historyManager := history.NewManager(sessionManager, checkpointManager, storage)

	// Get all sessions
	sessions, err := historyManager.GetAllSessions("", "")
	if err != nil {
		fmt.Printf("Error getting sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return
	}

	fmt.Println("Sessions:")
	fmt.Println()

	for _, sess := range sessions {
		status := "Ended"
		if sess.EndTime == nil {
			status = "Active"
		}

		fmt.Printf("%s  %s  %s  %s\n",
			sess.ID[:8],
			sess.Agent,
			sess.StartTime.Format("2006-01-02 15:04"),
			status)
	}
}

func handleContext() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Parse flags
	var task string
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--task":
			if i+1 < len(args) {
				task = args[i+1]
				i++
			}
		}
	}

	// Create managers
	sessionManager := session.NewManager(storage)
	checkpointManager := checkpoint.NewManager(storage)
	historyManager := history.NewManager(sessionManager, checkpointManager, storage)
	contextManager := agentcontext.NewManager(historyManager, checkpointManager, storage)

	// Compile context
	ctx, err := contextManager.Compile(repo, task)
	if err != nil {
		fmt.Printf("Error compiling context: %v\n", err)
		os.Exit(1)
	}

	// Format and print
	fmt.Println(contextManager.Format(ctx))
}

func handleDecision() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Get current session
	sessionManager := session.NewManager(storage)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Println("No active session. Run 'agent-ledger start' first.")
		os.Exit(1)
	}

	// Parse flags (simplified for MVP)
	var title, decision, rationale string
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--title":
			if i+1 < len(args) {
				title = args[i+1]
				i++
			}
		case "--decision":
			if i+1 < len(args) {
				decision = args[i+1]
				i++
			}
		case "--rationale":
			if i+1 < len(args) {
				rationale = args[i+1]
				i++
			}
		}
	}

	if title == "" || decision == "" {
		fmt.Println("Usage: agent-ledger decision --title <title> --decision <decision> [--rationale <rationale>]")
		os.Exit(1)
	}

	// Create decision
	eventsManager := events.NewManager(storage)
	_, err = eventsManager.CreateDecision(currentSession.ID, title, decision, rationale, []string{}, []string{})
	if err != nil {
		fmt.Printf("Error creating decision: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Decision recorded successfully")
}

func handleDiscovery() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Get current session
	sessionManager := session.NewManager(storage)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Println("No active session. Run 'agent-ledger start' first.")
		os.Exit(1)
	}

	// Parse flags (simplified for MVP)
	var title, finding string
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--title":
			if i+1 < len(args) {
				title = args[i+1]
				i++
			}
		case "--finding":
			if i+1 < len(args) {
				finding = args[i+1]
				i++
			}
		}
	}

	if title == "" || finding == "" {
		fmt.Println("Usage: agent-ledger discovery --title <title> --finding <finding>")
		os.Exit(1)
	}

	// Create discovery
	eventsManager := events.NewManager(storage)
	_, err = eventsManager.CreateDiscovery(currentSession.ID, title, finding, []string{}, []string{})
	if err != nil {
		fmt.Printf("Error creating discovery: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Discovery recorded successfully")
}

func handleFailure() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Get current session
	sessionManager := session.NewManager(storage)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Println("No active session. Run 'agent-ledger start' first.")
		os.Exit(1)
	}

	// Parse flags (simplified for MVP)
	var title, attempted, why, lessons string
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--title":
			if i+1 < len(args) {
				title = args[i+1]
				i++
			}
		case "--attempted":
			if i+1 < len(args) {
				attempted = args[i+1]
				i++
			}
		case "--why":
			if i+1 < len(args) {
				why = args[i+1]
				i++
			}
		case "--lessons":
			if i+1 < len(args) {
				lessons = args[i+1]
				i++
			}
		}
	}

	if title == "" || attempted == "" || why == "" {
		fmt.Println("Usage: agent-ledger failure --title <title> --attempted <approach> --why <reason> [--lessons <lessons>]")
		os.Exit(1)
	}

	// Create failure
	eventsManager := events.NewManager(storage)
	_, err = eventsManager.CreateFailure(currentSession.ID, title, attempted, why, lessons)
	if err != nil {
		fmt.Printf("Error creating failure: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Failure recorded successfully")
}

func handleConstraint() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Get current session
	sessionManager := session.NewManager(storage)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Println("No active session. Run 'agent-ledger start' first.")
		os.Exit(1)
	}

	// Parse flags (simplified for MVP)
	var title, constraint, reason string
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--title":
			if i+1 < len(args) {
				title = args[i+1]
				i++
			}
		case "--constraint":
			if i+1 < len(args) {
				constraint = args[i+1]
				i++
			}
		case "--reason":
			if i+1 < len(args) {
				reason = args[i+1]
				i++
			}
		}
	}

	if title == "" || constraint == "" {
		fmt.Println("Usage: agent-ledger constraint --title <title> --constraint <constraint> [--reason <reason>]")
		os.Exit(1)
	}

	// Create constraint
	eventsManager := events.NewManager(storage)
	_, err = eventsManager.CreateConstraint(currentSession.ID, title, constraint, reason)
	if err != nil {
		fmt.Printf("Error creating constraint: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Constraint recorded successfully")
}

func handleHandoff() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Get current session
	sessionManager := session.NewManager(storage)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Println("No active session. Run 'agent-ledger start' first.")
		os.Exit(1)
	}

	// Parse flags (simplified for MVP)
	var currentState, whatChanged string
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--state":
			if i+1 < len(args) {
				currentState = args[i+1]
				i++
			}
		case "--changed":
			if i+1 < len(args) {
				whatChanged = args[i+1]
				i++
			}
		}
	}

	if currentState == "" {
		currentState = "Working directory state preserved"
	}
	if whatChanged == "" {
		whatChanged = "General development work"
	}

	// Create handoff
	eventsManager := events.NewManager(storage)
	_, err = eventsManager.CreateHandoff(currentSession.ID, currentState, whatChanged, []string{}, []string{}, []string{}, []string{}, "", "", []string{}, []string{})
	if err != nil {
		fmt.Printf("Error creating handoff: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Handoff created successfully")
}

func handleExplain() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: agent-ledger explain <file>")
		os.Exit(1)
	}

	filePath := os.Args[2]

	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Create managers
	sessionManager := session.NewManager(storage)
	checkpointManager := checkpoint.NewManager(storage)
	historyManager := history.NewManager(sessionManager, checkpointManager, storage)
	contextManager := agentcontext.NewManager(historyManager, checkpointManager, storage)

	// Try to get active session for better context
	currentSession, err := sessionManager.GetCurrent()
	if err == nil {
		// Use session-aware explain
		explanation, err := contextManager.ExplainWithSession(filePath, currentSession.ID)
		if err != nil {
			fmt.Printf("Error explaining file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(explanation)
	} else {
		// Use basic explain
		explanation, err := contextManager.Explain(filePath)
		if err != nil {
			fmt.Printf("Error explaining file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(explanation)
	}
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func handleValidate() {
	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	st := storage.New(repo.Root)
	if !st.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	fmt.Println("Validating Agent Ledger...")
	fmt.Println()

	v := validator.NewValidator()

	// Validate sessions
	sessionDirs, err := st.ListDirectories("sessions")
	if err != nil {
		v.AddIssue(validator.LevelError, "sessions", fmt.Sprintf("Failed to list sessions: %v", err), "sessions/", "Check .agent/sessions directory permissions")
	} else {
		for _, sessionID := range sessionDirs {
			metadataPath := fmt.Sprintf("sessions/%s/metadata.json", sessionID)
			if !st.FileExists(metadataPath) {
				v.AddIssue(validator.LevelError, "session", fmt.Sprintf("Session %s missing metadata.json", sessionID), metadataPath, "The session folder exists but has no metadata")
			} else {
				var meta struct {
					ID        string  `json:"id"`
					Agent     string  `json:"agent,omitempty"`
					Model     string  `json:"model,omitempty"`
					Branch    string  `json:"branch"`
				}
				if readErr := st.ReadJSON(metadataPath, &meta); readErr == nil {
					v.ValidateSessionMetadata(meta.ID, meta.Agent, meta.Model, meta.Branch)
				}
			}
		}
	}

	// Validate semantic records
	decisions, _ := st.ListFiles("decisions")
	discoveries, _ := st.ListFiles("discoveries")
	failures, _ := st.ListFiles("failures")
	constraints, _ := st.ListFiles("constraints")

	for _, file := range decisions {
		path := "decisions/" + file
		content, readErr := st.ReadMarkdown(path)
		v.ValidateRecord("decision", file, file, content, path)
		v.ValidateFileFormat(path, content)
		if readErr != nil {
			v.AddIssue(validator.LevelError, "decision", fmt.Sprintf("Failed to read: %v", readErr), path, "Check file permissions")
		}
	}
	for _, file := range discoveries {
		path := "discoveries/" + file
		content, _ := st.ReadMarkdown(path)
		v.ValidateRecord("discovery", file, file, content, path)
		v.ValidateFileFormat(path, content)
	}
	for _, file := range failures {
		path := "failures/" + file
		content, _ := st.ReadMarkdown(path)
		v.ValidateRecord("failure", file, file, content, path)
		v.ValidateFileFormat(path, content)
	}
	for _, file := range constraints {
		path := "constraints/" + file
		content, _ := st.ReadMarkdown(path)
		v.ValidateRecord("constraint", file, file, content, path)
		v.ValidateFileFormat(path, content)
	}

	// Check checkpoint refs in Git
	refOutput, refErr := git.Command("show-ref")
	if refErr == nil {
		var checkpointRefs int
		for _, ref := range strings.Split(refOutput, "\n") {
			if strings.Contains(ref, "refs/agents/sessions/") {
				checkpointRefs++
			}
		}
		fmt.Printf("Git checkpoint refs: %d\n", checkpointRefs)
	}

	// Overall counts
	totalRecords := len(decisions) + len(discoveries) + len(failures) + len(constraints)
	comp := validator.RunComprehensiveValidation(len(sessionDirs), totalRecords)
	for _, issue := range comp.Issues {
		v.AddIssue(issue.Level, issue.Category, issue.Message, issue.Path, issue.Suggestion)
	}

	// Print results
	result := v.Result()

	fmt.Printf("Sessions: %d  |  Records: %d decisions, %d discoveries, %d failures, %d constraints\n",
		len(sessionDirs), len(decisions), len(discoveries), len(failures), len(constraints))
	fmt.Println()

	if len(result.Issues) == 0 {
		fmt.Println("✓ Validation passed — no issues found")
		return
	}

	// Print issues grouped by level
	for _, issue := range result.Issues {
		var prefix string
		switch issue.Level {
		case validator.LevelError:
			prefix = "[ERROR]"
		case validator.LevelWarning:
			prefix = "[WARN] "
		case validator.LevelInfo:
			prefix = "[INFO] "
		}
		fmt.Printf("%s %s: %s\n", prefix, issue.Category, issue.Message)
		if issue.Suggestion != "" {
			fmt.Printf("         → %s\n", issue.Suggestion)
		}
	}

	fmt.Println()
	if result.Valid {
		fmt.Printf("Validation completed in %s — %d warning(s), 0 errors\n", result.Duration, result.Warnings)
	} else {
		fmt.Printf("Validation completed in %s — %d error(s), %d warning(s)\n", result.Duration, result.Errors, result.Warnings)
		os.Exit(1)
	}
}

func handleUI() {
	// Check for help flag
	args := os.Args[2:]
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			fmt.Println("agent-ledger ui - Launch local web UI")
			fmt.Println()
			fmt.Println("Usage:")
			fmt.Println("  agent-ledger ui [--port <port>] [--host <host>] [--open]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --port <port>  Port to run the UI server on (default: 5173)")
			fmt.Println("  --host <host>  Host to bind the UI server to (default: localhost)")
			fmt.Println("  --open         Automatically open browser after starting server")
			return
		}
	}

	// Check if we're in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		fmt.Printf("Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Check if agent ledger exists
	storage := storage.New(repo.Root)
	if !storage.Exists() {
		fmt.Println("Agent ledger not initialized. Run 'agent-ledger init' first.")
		os.Exit(1)
	}

	// Parse flags
	port := 5173
	host := "localhost"
	openBrowser := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &port)
				i++
			}
		case "--host":
			if i+1 < len(args) {
				host = args[i+1]
				i++
			}
		case "--open":
			openBrowser = true
		}
	}

	// Start the UI server
	server := ui.NewServer(repo, storage, port, host, Version)
	if err := server.Start(); err != nil {
		fmt.Printf("Error starting UI server: %v\n", err)
		os.Exit(1)
	}

	// Auto-open browser if requested
	if openBrowser {
		url := fmt.Sprintf("http://%s:%d", host, port)
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		case "darwin":
			cmd = exec.Command("open", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		_ = cmd.Start() // Ignore errors, browser opening is optional
	}
}

func handleMCP() {
	mcpinternal.RunWithLogging(Version)
}
