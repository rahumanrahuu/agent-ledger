package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
	agentledger "agent-ledger/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var Version = "dev"

func printMCPUsage() {
	fmt.Println("ledger-mcp - Model Context Protocol (MCP) server for Agent Ledger")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ledger-mcp               Start MCP server over stdio")
	fmt.Println("  ledger-mcp --help, -h    Show this help message")
	fmt.Println("  ledger-mcp --version, -v Show version")
	fmt.Println()
	fmt.Println("The MCP server connects AI coding agents to Agent Ledger over stdio.")
	fmt.Println("It automatically locates the nearest Git repository root by walking upward from")
	fmt.Println("the current working directory.")
}

func main() {
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if arg == "--help" || arg == "-h" || arg == "help" {
				printMCPUsage()
				return
			}
			if arg == "--version" || arg == "-v" || arg == "version" {
				fmt.Printf("ledger-mcp %s\n", Version)
				return
			}
		}
	}

	// Locate repository root by walking upward from current working directory
	repoRoot, err := repository.FindRepositoryRoot()
	if err != nil {
		log.Fatalf("Error: %v - agent-ledger requires a git repository to work", err)
	}

	// Change working directory to repository root so all operations are anchored properly
	if err := os.Chdir(repoRoot); err != nil {
		log.Fatalf("Error switching to repository root: %v", err)
	}

	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		log.Fatalf("Error detecting repository: %v", err)
	}

	// Initialize storage
	st := storage.New(repo.Root)
	if !st.Exists() {
		log.Fatal("Agent ledger not initialized in this repository. Run 'agent-ledger init' first.")
	}

	// Create MCP manager
	mcpManager, err := agentledger.NewManager(st)
	if err != nil {
		log.Fatalf("Error creating MCP manager: %v", err)
	}

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "agent-ledger",
		Version: Version,
	}, nil)

	// Register resources
	if err := mcpManager.RegisterResources(server); err != nil {
		log.Fatalf("Error registering resources: %v", err)
	}

	// Register tools
	if err := mcpManager.RegisterTools(server); err != nil {
		log.Fatalf("Error registering tools: %v", err)
	}

	// Run server over stdin/stdout
	ctx := context.Background()
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}
