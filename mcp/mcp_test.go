package mcp

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
)

func TestMCPServerInitialization(t *testing.T) {
	// Create a temporary test repository
	tmpDir, err := os.MkdirTemp("", "agent-ledger-mcp-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Initialize a git repository
	// For now, we'll skip the actual git initialization since this is just testing MCP schema validation
	// and instead create the agent ledger structure directly
	
	// Initialize agent ledger storage
	st := storage.New(tmpDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	
	// Create MCP manager
	mcpManager, err := NewManager(st)
	if err != nil {
		t.Fatalf("Failed to create MCP manager: %v", err)
	}
	
	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "agent-ledger",
		Version: "1.0.0",
	}, nil)
	
	// Register resources
	if err := mcpManager.RegisterResources(server); err != nil {
		t.Fatalf("Failed to register resources: %v", err)
	}
	
	// Register tools - this should not panic with missing input schemas
	if err := mcpManager.RegisterTools(server); err != nil {
		t.Fatalf("Failed to register tools: %v", err)
	}
	
	// Test that the server was created successfully
	if server == nil {
		t.Fatal("Server should not be nil after successful initialization")
	}
}

func TestToolSchemas(t *testing.T) {
	// Verify that our input schemas are valid
	t.Run("emptyInputSchema should be valid", func(t *testing.T) {
		var schema map[string]interface{}
		if err := json.Unmarshal(emptyInputSchema, &schema); err != nil {
			t.Fatalf("Failed to unmarshal emptyInputSchema: %v", err)
		}
		if schema["type"] != "object" {
			t.Errorf("emptyInputSchema type should be 'object', got %v", schema["type"])
		}
	})
	
	t.Run("contextInputSchema should be valid", func(t *testing.T) {
		var schema map[string]interface{}
		if err := json.Unmarshal(contextInputSchema, &schema); err != nil {
			t.Fatalf("Failed to unmarshal contextInputSchema: %v", err)
		}
		if schema["type"] != "object" {
			t.Errorf("contextInputSchema type should be 'object', got %v", schema["type"])
		}
		props, ok := schema["properties"].(map[string]interface{})
		if !ok {
			t.Error("contextInputSchema properties should be a map")
		}
		if _, ok := props["task"]; !ok {
			t.Error("contextInputSchema should have 'task' property")
		}
	})
	
	t.Run("historyInputSchema should be valid", func(t *testing.T) {
		var schema map[string]interface{}
		if err := json.Unmarshal(historyInputSchema, &schema); err != nil {
			t.Fatalf("Failed to unmarshal historyInputSchema: %v", err)
		}
		if schema["type"] != "object" {
			t.Errorf("historyInputSchema type should be 'object', got %v", schema["type"])
		}
		props, ok := schema["properties"].(map[string]interface{})
		if !ok {
			t.Error("historyInputSchema properties should be a map")
		}
		if _, ok := props["session_id"]; !ok {
			t.Error("historyInputSchema should have 'session_id' property")
		}
	})
}

func TestToolHandlersExist(t *testing.T) {
	// Create a temporary test repository
	tmpDir, err := os.MkdirTemp("", "agent-ledger-mcp-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Initialize agent ledger storage
	st := storage.New(tmpDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	
	// Create MCP manager
	mcpManager, err := NewManager(st)
	if err != nil {
		t.Fatalf("Failed to create MCP manager: %v", err)
	}
	
	// Verify that the manager was created successfully
	if mcpManager == nil {
		t.Error("MCP manager should not be nil")
	}
	
	// Verify that internal managers are initialized
	if mcpManager.sessionManager == nil {
		t.Error("sessionManager should be initialized")
	}
	if mcpManager.checkpointManager == nil {
		t.Error("checkpointManager should be initialized")
	}
	if mcpManager.historyManager == nil {
		t.Error("historyManager should be initialized")
	}
	if mcpManager.contextManager == nil {
		t.Error("contextManager should be initialized")
	}
	if mcpManager.storage == nil {
		t.Error("storage should be initialized")
	}
}

func TestToolRegistrationInRepository(t *testing.T) {
	// This test requires a real git repository, so we skip it if not in one
	if err := repository.MustBeInRepository(); err != nil {
		t.Skip("Not in a git repository, skipping integration test")
	}
	
	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		t.Fatalf("Failed to detect repository: %v", err)
	}
	
	// Initialize agent ledger storage
	st := storage.New(repo.Root)
	if !st.Exists() {
		t.Skip("Agent ledger not initialized, skipping integration test")
	}
	
	// Create MCP manager
	mcpManager, err := NewManager(st)
	if err != nil {
		t.Fatalf("Failed to create MCP manager: %v", err)
	}
	
	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "agent-ledger",
		Version: "1.0.0",
	}, nil)
	
	// Register tools - this should not panic with missing input schemas
	if err := mcpManager.RegisterTools(server); err != nil {
		t.Fatalf("Failed to register tools: %v", err)
	}
	
	// Test tool handlers with a context
	ctx := context.Background()
	
	// Test that handleCheckpoint can be called (it may fail due to no active session, but shouldn't panic)
	t.Run("handleCheckpoint handles no session gracefully", func(t *testing.T) {
		req := &mcp.CallToolRequest{}
		result, err := mcpManager.handleCheckpoint(ctx, req)
		if err != nil {
			t.Errorf("handleCheckpoint should not error, got: %v", err)
		}
		if result == nil {
			t.Error("handleCheckpoint should return a result")
		}
	})
	
	// Test that handleGetContext can be called
	t.Run("handleGetContext works", func(t *testing.T) {
		req := &mcp.CallToolRequest{}
		result, err := mcpManager.handleGetContext(ctx, req)
		if err != nil {
			t.Errorf("handleGetContext should not error, got: %v", err)
		}
		if result == nil {
			t.Error("handleGetContext should return a result")
		}
	})
	
	// Test that handleGetHistory can be called
	t.Run("handleGetHistory works", func(t *testing.T) {
		req := &mcp.CallToolRequest{}
		result, err := mcpManager.handleGetHistory(ctx, req)
		if err != nil {
			t.Errorf("handleGetHistory should not error, got: %v", err)
		}
		if result == nil {
			t.Error("handleGetHistory should return a result")
		}
	})
	
	// If we get here, the server was created successfully without panicking
	t.Log("MCP server initialization succeeded with valid input schemas")
}
