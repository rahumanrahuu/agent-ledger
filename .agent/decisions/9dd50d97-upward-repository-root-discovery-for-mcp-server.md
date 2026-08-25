# Upward repository root discovery for MCP server

**Decision:** MCP server walks upward from cwd to find repository root instead of requiring explicit cwd or hardcoded path. Uses FindRepositoryRoot() which calls IsRepositoryInDir() at each parent.

**Rationale:** MCP hosts may launch ledger-mcp from different working directories (e.g., the user's home directory). Upward walking ensures the server finds the correct repository without hardcoded paths in mcp_config.json.


*Session: 4417de2f-d064-4f67-b491-9a71f20fa80b*
*Created: 2026-08-25T16:43:38Z*
