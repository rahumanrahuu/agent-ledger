package mcpinternal

import (
	"context"
	"fmt"
	"log"
	"os"

	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
	agentledger "agent-ledger/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Run starts the MCP server over stdio
func Run(ctx context.Context, version string) error {
	// Locate repository root by walking upward from current working directory
	repoRoot, err := repository.FindRepositoryRoot()
	if err != nil {
		return fmt.Errorf("error: %v - agent-ledger requires a git repository to work", err)
	}

	// Change working directory to repository root so all operations are anchored properly
	if err := os.Chdir(repoRoot); err != nil {
		return fmt.Errorf("error switching to repository root: %v", err)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		return fmt.Errorf("error detecting repository: %v", err)
	}

	// Initialize storage
	st := storage.New(repo.Root)
	if !st.Exists() {
		return fmt.Errorf("agent ledger not initialized in this repository. Run 'agent-ledger init' first.")
	}

	// Create MCP manager
	mcpManager, err := agentledger.NewManager(st)
	if err != nil {
		return fmt.Errorf("error creating MCP manager: %v", err)
	}

	// Create MCP server
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "agent-ledger",
		Version: version,
	}, nil)

	// Register resources
	if err := mcpManager.RegisterResources(server); err != nil {
		return fmt.Errorf("error registering resources: %v", err)
	}

	// Register tools
	if err := mcpManager.RegisterTools(server); err != nil {
		return fmt.Errorf("error registering tools: %v", err)
	}

	// Run server over stdin/stdout
	if err := server.Run(ctx, &mcpsdk.StdioTransport{}); err != nil {
		return fmt.Errorf("error running server: %v", err)
	}

	return nil
}

// RunWithLogging runs the MCP server with logging for standalone binary usage
func RunWithLogging(version string) {
	// Check for help/version flags
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if arg == "--help" || arg == "-h" || arg == "help" {
				printMCPUsage()
				return
			}
			if arg == "--version" || arg == "-v" || arg == "version" {
				fmt.Printf("agent-ledger mcp %s\n", version)
				return
			}
		}
	}

	ctx := context.Background()
	if err := Run(ctx, version); err != nil {
		log.Fatal(err)
	}
}

func printMCPUsage() {
	fmt.Println("agent-ledger mcp - Model Context Protocol (MCP) server for Agent Ledger")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  agent-ledger mcp               Start MCP server over stdio")
	fmt.Println("  agent-ledger mcp --help, -h    Show this help message")
	fmt.Println("  agent-ledger mcp --version, -v Show version")
	fmt.Println()
	fmt.Println("The MCP server connects AI coding agents to Agent Ledger over stdio.")
	fmt.Println("It automatically locates the nearest Git repository root by walking upward from")
	fmt.Println("the current working directory.")
}
