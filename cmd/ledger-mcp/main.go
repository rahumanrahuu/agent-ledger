package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
	agentledgervmcp "agent-ledger/mcp"
)

func main() {
	// Check if we are in a git repository
	if err := repository.MustBeInRepository(); err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get repository root
	repo, err := repository.Detect()
	if err != nil {
		log.Fatalf("Error detecting repository: %v", err)
	}

	// Initialize storage
	st := storage.New(repo.Root)
	if err := st.Initialize(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create MCP manager
	manager, err := agentledgervmcp.NewManager(st)
	if err != nil {
		log.Fatalf("Failed to create MCP manager: %v", err)
	}

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{Name: "agent-ledger", Version: "1.0.0"}, nil)
	
	// Register resources
	if err := manager.RegisterResources(server); err != nil {
		log.Fatalf("Failed to register resources: %v", err)
	}

	// Register tools
	if err := manager.RegisterTools(server); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Run server on stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
