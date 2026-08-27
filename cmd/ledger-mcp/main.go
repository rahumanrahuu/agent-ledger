package main

import (
	"fmt"
	"os"

	mcpinternal "agent-ledger/internal/mcp"
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
	fmt.Println()
	fmt.Println("Note: ledger-mcp is deprecated. Use 'agent-ledger mcp' instead.")
}

func main() {
	// Check for help/version flags first
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

	// Use the shared MCP logic
	mcpinternal.RunWithLogging(Version)
}
