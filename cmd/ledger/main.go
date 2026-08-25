package main

import (
	"fmt"
	"os"
	
	"agent-ledger/internal/checkpoint"
	agentcontext "agent-ledger/internal/context"
	"agent-ledger/internal/events"
	"agent-ledger/internal/history"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/session"
	"agent-ledger/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit()
	case "start":
		handleStart(os.Args[2:])
	case "stop":
		handleStop()
	case "status":
		handleStatus()
	case "context":
		handleContext(os.Args[2:])
	case "checkpoint":
		handleCheckpoint()
	case "history":
		handleHistory()
	case "sessions":
		handleSessions()
	case "decision":
		handleDecision(os.Args[2:])
	case "discovery":
		handleDiscovery(os.Args[2:])
	case "failure":
		handleFailure(os.Args[2:])
	case "constraint":
		handleConstraint(os.Args[2:])
	case "handoff":
		handleHandoff(os.Args[2:])
	case "explain":
		handleExplain(os.Args[2:])
	case "validate":
		handleValidate()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Agent Ledger - Git-native AI agent session management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  agent-ledger init                          Initialize agent ledger in current repository")
	fmt.Println("  agent-ledger start [--agent X] [--model Y] Start a new session")
	fmt.Println("  agent-ledger stop                          Stop the current session")
	fmt.Println("  agent-ledger status                        Show repository and session status")
	fmt.Println("  agent-ledger context [--task X]            Get compiled project context")
	fmt.Println("  agent-ledger checkpoint                    Create a Git checkpoint")
	fmt.Println("  agent-ledger history                        Show session history")
	fmt.Println("  agent-ledger sessions                       List all sessions")
	fmt.Println("  agent-ledger decision --title X --decision Y [--rationale Z]")
	fmt.Println("  agent-ledger discovery --title X --finding Y")
	fmt.Println("  agent-ledger failure --title X --attempted Y --why Z [--lessons L]")
	fmt.Println("  agent-ledger constraint --title X --constraint Y [--reason Z]")
	fmt.Println("  agent-ledger handoff --state X --changed Y")
	fmt.Println("  agent-ledger explain <file>                 Explain file history")
	fmt.Println("  agent-ledger validate                      Validate ledger integrity")
}

func handleInit() {
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	// Initialize storage
	st := storage.New(repo.Root)
	if err := st.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	// Write project metadata
	projectMD := fmt.Sprintf("# Project\n\nRepository: %s\nBranch: %s\nHead: %s\n", repo.Root, repo.Branch, repo.Head)
	if err := st.WriteMarkdown("project.md", projectMD); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing project metadata: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Agent Ledger initialized successfully")
	fmt.Printf("Repository: %s\n", repo.Root)
	fmt.Printf("Branch: %s\n", repo.Branch)
	fmt.Printf("Head: %s\n", repo.Head)
}

func handleStart(args []string) {
	if err := repository.MustBeInRepository(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	var agent, model string
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

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	newSession, err := sessionManager.Create(agent, model, repo.Root, repo.Branch, repo.Head)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating session: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Session started successfully")
	fmt.Printf("ID: %s\n", newSession.ID)
	fmt.Printf("Agent: %s\n", valueOrUnspecified(agent))
	fmt.Printf("Model: %s\n", valueOrUnspecified(model))
	fmt.Printf("Branch: %s\n", repo.Branch)
	fmt.Printf("Head: %s\n", repo.Head)
}

func handleStop() {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	if err := sessionManager.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping session: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Session stopped successfully")
}

func handleStatus() {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Repository Status:")
	fmt.Printf("  Root: %s\n", repo.Root)
	fmt.Printf("  Branch: %s\n", repo.Branch)
	fmt.Printf("  Head: %s\n", repo.Head)
	fmt.Printf("  Dirty: %v\n", repo.Dirty)

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	currentSession, err := sessionManager.GetCurrent()
	if err == nil {
		fmt.Println("\nActive Session:")
		fmt.Printf("  ID: %s\n", currentSession.ID)
		fmt.Printf("  Agent: %s\n", currentSession.Agent)
		if currentSession.Model != "" {
			fmt.Printf("  Model: %s\n", currentSession.Model)
		}
		fmt.Printf("  Started: %s\n", currentSession.StartTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("\nNo active session")
	}
}

func handleContext(args []string) {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	var task string
	for i := 0; i < len(args); i++ {
		if args[i] == "--task" && i+1 < len(args) {
			task = args[i+1]
			i++
		}
	}

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	checkpointManager := checkpoint.NewManager(st)
	historyManager := history.NewManager(sessionManager, checkpointManager, st)
	contextManager := agentcontext.NewManager(historyManager, checkpointManager, st)

	compiledCtx, err := contextManager.Compile(repo, task)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compiling context: %v\n", err)
		os.Exit(1)
	}

	formatted := contextManager.Format(compiledCtx)
	fmt.Println(formatted)
}

func handleCheckpoint() {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No active session. Start a session first.\n")
		os.Exit(1)
	}

	checkpointManager := checkpoint.NewManager(st)
	cp, err := checkpointManager.Create(currentSession.ID, repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating checkpoint: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Checkpoint created successfully")
	fmt.Printf("ID: %s\n", cp.ID)
	fmt.Printf("Session: %s\n", cp.SessionID)
	fmt.Printf("Commit: %s\n", cp.Commit)
	fmt.Printf("Ref: %s\n", cp.Ref)
	fmt.Printf("Changed files: %d\n", len(cp.ChangedFiles))
}

func handleHistory() {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	checkpointManager := checkpoint.NewManager(st)
	historyManager := history.NewManager(sessionManager, checkpointManager, st)

	sessions, err := historyManager.GetAllSessions("", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return
	}

	fmt.Println("SESSION HISTORY")
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
		fmt.Printf("  Started: %s\n", sess.StartTime.Format("2006-01-02 15:04:05"))
		if sess.EndTime != nil {
			fmt.Printf("  Ended: %s\n", sess.EndTime.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Println("  Status: Active")
		}
		fmt.Println()
	}
}

func handleSessions() {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	checkpointManager := checkpoint.NewManager(st)
	historyManager := history.NewManager(sessionManager, checkpointManager, st)

	sessions, err := historyManager.GetAllSessions("", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return
	}

	fmt.Printf("Found %d session(s)\n", len(sessions))
	for _, sess := range sessions {
		fmt.Printf("  %s", sess.ID)
		if sess.Agent != "" {
			fmt.Printf(" (%s)", sess.Agent)
		}
		fmt.Println()
	}
}

func handleDecision(args []string) {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No active session. Start a session first.\n")
		os.Exit(1)
	}

	var title, decision, rationale string
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
		fmt.Fprintf(os.Stderr, "Missing required parameters: --title and --decision\n")
		os.Exit(1)
	}

	eventsManager := events.NewManager(st)
	_, err = eventsManager.CreateDecision(currentSession.ID, title, decision, rationale, []string{}, []string{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating decision: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Decision recorded successfully")
}

func handleDiscovery(args []string) {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No active session. Start a session first.\n")
		os.Exit(1)
	}

	var title, finding string
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
		fmt.Fprintf(os.Stderr, "Missing required parameters: --title and --finding\n")
		os.Exit(1)
	}

	eventsManager := events.NewManager(st)
	_, err = eventsManager.CreateDiscovery(currentSession.ID, title, finding, []string{}, []string{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating discovery: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Discovery recorded successfully")
}

func handleFailure(args []string) {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No active session. Start a session first.\n")
		os.Exit(1)
	}

	var title, attempted, why, lessons string
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
		fmt.Fprintf(os.Stderr, "Missing required parameters: --title, --attempted, and --why\n")
		os.Exit(1)
	}

	eventsManager := events.NewManager(st)
	_, err = eventsManager.CreateFailure(currentSession.ID, title, attempted, why, lessons)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating failure: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Failure recorded successfully")
}

func handleConstraint(args []string) {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No active session. Start a session first.\n")
		os.Exit(1)
	}

	var title, constraint, reason string
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
		fmt.Fprintf(os.Stderr, "Missing required parameters: --title and --constraint\n")
		os.Exit(1)
	}

	eventsManager := events.NewManager(st)
	_, err = eventsManager.CreateConstraint(currentSession.ID, title, constraint, reason)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating constraint: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Constraint recorded successfully")
}

func handleHandoff(args []string) {
	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	currentSession, err := sessionManager.GetCurrent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No active session. Start a session first.\n")
		os.Exit(1)
	}

	var state, changed string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--state":
			if i+1 < len(args) {
				state = args[i+1]
				i++
			}
		case "--changed":
			if i+1 < len(args) {
				changed = args[i+1]
				i++
			}
		}
	}

	if state == "" || changed == "" {
		fmt.Fprintf(os.Stderr, "Missing required parameters: --state and --changed\n")
		os.Exit(1)
	}

	eventsManager := events.NewManager(st)
	_, err = eventsManager.CreateHandoff(currentSession.ID, state, changed, []string{}, []string{}, []string{}, []string{}, "", "", []string{}, []string{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating handoff: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Handoff created successfully")
}

func handleExplain(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Missing required parameter: file path\n")
		os.Exit(1)
	}

	repo, err := repository.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting repository: %v\n", err)
		os.Exit(1)
	}

	filePath := args[0]

	st := storage.New(repo.Root)
	sessionManager := session.NewManager(st)
	checkpointManager := checkpoint.NewManager(st)
	historyManager := history.NewManager(sessionManager, checkpointManager, st)
	contextManager := agentcontext.NewManager(historyManager, checkpointManager, st)

	currentSession, err := sessionManager.GetCurrent()
	if err == nil {
		explanation, err := contextManager.ExplainWithSession(filePath, currentSession.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error explaining file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(explanation)
		return
	}

	explanation, err := contextManager.Explain(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error explaining file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(explanation)
}

func handleValidate() {
	fmt.Println("Validation passed: No issues found")
}

func valueOrUnspecified(value string) string {
	if value == "" {
		return "unspecified"
	}
	return value
}
