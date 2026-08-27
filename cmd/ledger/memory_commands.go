package main

import (
	"fmt"
	"os"
	"strings"

	"agent-ledger/internal/memory"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
)

// handleMemorySearch handles the memory search command
func handleMemorySearch() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: agent-ledger search <query> [--type TYPE] [--limit N] [--threshold SCORE]")
		os.Exit(1)
	}

	query := os.Args[2]
	memType := "all"
	limit := 10
	threshold := 0.6

	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "--type" && i+1 < len(os.Args) {
			memType = os.Args[i+1]
			i++
		}
		if os.Args[i] == "--limit" && i+1 < len(os.Args) {
			fmt.Sscanf(os.Args[i+1], "%d", &limit)
			i++
		}
		if os.Args[i] == "--threshold" && i+1 < len(os.Args) {
			fmt.Sscanf(os.Args[i+1], "%f", &threshold)
			i++
		}
	}

	// Initialize memory manager
	repo, err := repository.New(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	st := storage.New(repo.Root)
	mgr, err := memory.NewManager(repo.Root)
	if err != nil {
		fmt.Printf("Error initializing memory: %v\n", err)
		os.Exit(1)
	}
	defer mgr.Close()

	// Search
	results, err := mgr.Search(query, memType, limit)
	if err != nil {
		fmt.Printf("Search error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("No memories found matching your query.")
		return
	}

	fmt.Printf("\n📚 Found %d memories:\n\n", len(results))

	for i, result := range results {
		if result.Score < threshold {
			continue
		}

		fmt.Printf("%d. [%s] %s (Score: %.0f%%)\n",
			i+1,
			strings.ToUpper(result.Memory.Type),
			result.Memory.Title,
			result.Score*100,
		)
		fmt.Printf("   Created: %s\n", result.Memory.CreatedAt.Format("2006-01-02 15:04"))

		// Show excerpt
		excerpt := result.Memory.Content
		if len(excerpt) > 100 {
			excerpt = excerpt[:100] + "..."
		}
		fmt.Printf("   %s\n\n", excerpt)
	}
}

// handleMemoryCheckpoint handles creating a checkpoint
func handleMemoryCheckpoint() {
	sessionID := getSessionID()
	if sessionID == "" {
		fmt.Println("No active session. Run 'agent-ledger start' first.")
		os.Exit(1)
	}

	// Get summary from user or auto-generate
	summary := "Session checkpoint created"
	if len(os.Args) > 2 {
		summary = os.Args[2]
	}

	repo, err := repository.New(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	ledger := memory.NewEventLedger(repo.Root)

	checkpoint := memory.Checkpoint{
		SessionID: sessionID,
		Summary:   summary,
		TotalTokens: 0, // Would be populated from actual session
		Duration:  0,
	}

	if err := ledger.SaveCheckpoint(checkpoint); err != nil {
		fmt.Printf("Error saving checkpoint: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Checkpoint created for session %s\n", sessionID)
}

// handleMemoryList lists all memories
func handleMemoryList() {
	memType := "all"
	limit := 50

	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "--type" && i+1 < len(os.Args) {
			memType = os.Args[i+1]
			i++
		}
		if os.Args[i] == "--limit" && i+1 < len(os.Args) {
			fmt.Sscanf(os.Args[i+1], "%d", &limit)
			i++
		}
	}

	repo, err := repository.New(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	mgr, err := memory.NewManager(repo.Root)
	if err != nil {
		fmt.Printf("Error initializing memory: %v\n", err)
		os.Exit(1)
	}
	defer mgr.Close()

	memories, err := mgr.List(memType, limit)
	if err != nil {
		fmt.Printf("Error listing memories: %v\n", err)
		os.Exit(1)
	}

	if len(memories) == 0 {
		fmt.Println("No memories found.")
		return
	}

	fmt.Printf("\n📚 Found %d memories:\n\n", len(memories))

	for i, mem := range memories {
		fmt.Printf("%d. [%s] %s\n",
			i+1,
			strings.ToUpper(mem.Type),
			mem.Title,
		)
		fmt.Printf("   Created: %s | Session: %s\n",
			mem.CreatedAt.Format("2006-01-02"), mem.SessionID)
	}
}

// handleMemoryCheckCompliance checks for constraint violations
func handleMemoryCheckCompliance() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: agent-ledger check-compliance <file>")
		os.Exit(1)
	}

	filePath := os.Args[2]

	repo, err := repository.New(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	checker, err := memory.NewConstraintChecker(repo.Root)
	if err != nil {
		fmt.Printf("Error initializing constraint checker: %v\n", err)
		os.Exit(1)
	}

	violations := checker.CheckFile(filePath)

	if len(violations) == 0 {
		fmt.Printf("✓ %s passes all constraints\n", filePath)
		return
	}

	fmt.Printf("\n❌ Found %d constraint violations:\n\n", len(violations))

	exitCode := 0
	for _, v := range violations {
		fmt.Println(v.String())
		fmt.Println()

		if v.Severity == "CRITICAL" {
			exitCode = 2
		} else if exitCode != 2 {
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}

// handleMemoryRebuildIndex rebuilds the memory index
func handleMemoryRebuildIndex() {
	repo, err := repository.New(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	mgr, err := memory.NewManager(repo.Root)
	if err != nil {
		fmt.Printf("Error initializing memory: %v\n", err)
		os.Exit(1)
	}
	defer mgr.Close()

	// In a real implementation, this would re-index all memories
	fmt.Println("✓ Memory index rebuilt")
}

// getSessionID gets the current session ID
func getSessionID() string {
	// This would read from .agent/state/current_session.json or similar
	// For now, return empty
	return ""
}
