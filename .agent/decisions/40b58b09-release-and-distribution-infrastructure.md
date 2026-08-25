# Release and distribution infrastructure

**Decision:** Implemented GoReleaser + GitHub Actions + install.sh + install.ps1 for zero-Go-required installation across macOS/Linux/Windows (amd64+arm64). Binary names standardized to agent-ledger and ledger-mcp.

**Rationale:** Users should install Agent Ledger from GitHub Releases without Go or source cloning. GoReleaser produces consistent cross-platform archives with ldflags-injected versions.


*Session: 4417de2f-d064-4f67-b491-9a71f20fa80b*
*Created: 2026-08-25T16:43:28Z*
