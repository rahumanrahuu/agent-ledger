# Agent Ledger

A local-first, Git-native system that preserves development history and context across AI coding-agent sessions. Different agents can continue work without losing previous decisions, discoveries, failures, constraints, or code history.

Agent Ledger does **not** provide its own AI model or coding agent. Existing coding agents (Claude Code, Cursor, Copilot, Gemini, etc.) provide the reasoning. Agent Ledger provides persistent development context and history.

---

## Installation

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/rahumanrahuu/agent-ledger/main/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/rahumanrahuu/agent-ledger/main/install.ps1 | iex
```

### What gets installed

Two executables are placed in a user-local directory:

| Binary | Purpose |
|---|---|
| `agent-ledger` | CLI for session management, checkpoints, context, and semantic records |
| `ledger-mcp` | MCP server (stdio) that connects AI coding agents to Agent Ledger |

**Install locations:**

- **macOS / Linux**: `~/.local/bin`
- **Windows**: `%LOCALAPPDATA%\Programs\agent-ledger`

No `sudo` or Administrator privileges are required.

### Supported platforms

| OS | Architectures |
|---|---|
| macOS | arm64 (Apple Silicon), amd64 (Intel) |
| Linux | arm64, amd64 |
| Windows | arm64, amd64 |

### Installing a specific version

**Unix:**
```bash
VERSION=v0.2.1 curl -fsSL https://raw.githubusercontent.com/rahumanrahuu/agent-ledger/main/install.sh | sh
```

**Windows PowerShell:**
```powershell
$env:VERSION="v0.2.1"; irm https://raw.githubusercontent.com/rahumanrahuu/agent-ledger/main/install.ps1 | iex
```

### Verify installation

```bash
agent-ledger --help
agent-ledger --version
ledger-mcp --help
```

### PATH setup

If your shell reports `command not found` after installation, add the install directory to your PATH:

**macOS / Linux (add to `~/.bashrc`, `~/.zshrc`, or `~/.profile`):**
```bash
export PATH="$HOME/.local/bin:$PATH"
```

**Windows PowerShell (permanent for current user):**
```powershell
[System.Environment]::SetEnvironmentVariable('Path', $env:Path + ';' + "$env:LOCALAPPDATA\Programs\agent-ledger", 'User')
```

### Uninstall

Remove the two binaries from the install directory:

```bash
# macOS / Linux
rm ~/.local/bin/agent-ledger ~/.local/bin/ledger-mcp

# Windows PowerShell
Remove-Item "$env:LOCALAPPDATA\Programs\agent-ledger\agent-ledger.exe"
Remove-Item "$env:LOCALAPPDATA\Programs\agent-ledger\ledger-mcp.exe"
```

### Building from source

Requires Go 1.21+ and Git.

```bash
git clone https://github.com/rahumanrahuu/agent-ledger.git
cd agent-ledger
go build -o agent-ledger ./cmd/ledger
go build -o ledger-mcp ./cmd/ledger-mcp
```

---

## Getting Started

Agent Ledger requires a Git repository to operate.

```bash
# 1. Enter your project
cd my-project

# 2. Initialize Agent Ledger
agent-ledger init

# 3. Start a session (agent and model identifiers are arbitrary and optional)
agent-ledger start --agent cursor --model gpt-4

# 4. Work normally — make code changes, run tests, etc.

# 5. Create a checkpoint of the current working tree
agent-ledger checkpoint

# 6. Record what you learned
agent-ledger decision --title "Use PostgreSQL" --decision "Chose PostgreSQL over MongoDB" --rationale "ACID compliance required"
agent-ledger discovery --title "Rate limit needed" --finding "API has no rate limiting"

# 7. Create a handoff for the next agent
agent-ledger handoff --state "Auth system complete" --changed "Implemented JWT authentication"

# 8. Stop the session
agent-ledger stop
```

A subsequent agent reconstructs context:

```bash
agent-ledger start --agent claude-code
agent-ledger context
# → Returns architecture, decisions, discoveries, failures, constraints, and latest handoff
```

---

## CLI Reference

```
agent-ledger init                Initialize .agent/ directory in current repository
agent-ledger status              Show repository and ledger status
agent-ledger start [flags]       Start a session (--agent <name> --model <name>)
agent-ledger stop                Stop the current session
agent-ledger checkpoint          Create an index-preserving checkpoint
agent-ledger context [--task T]  Compile onboarding context (optionally task-specific)
agent-ledger decision            Record a decision (--title, --decision, --rationale)
agent-ledger discovery           Record a discovery (--title, --finding)
agent-ledger failure             Record a failure (--title, --attempted, --why, --lessons)
agent-ledger constraint          Record a constraint (--title, --constraint, --reason)
agent-ledger handoff             Create a handoff document (--state, --changed)
agent-ledger explain <file>      Explain development history of a file
agent-ledger history             Show session history with details
agent-ledger sessions            List all sessions
agent-ledger validate            Validate ledger integrity
agent-ledger --help              Show help
agent-ledger --version           Show version
```

---

## MCP Setup

`ledger-mcp` is a stdio-based Model Context Protocol (MCP) server. MCP clients launch it directly — you do not need to start it manually in a terminal.

The MCP server automatically locates the Git repository root by walking upward from its current working directory. No hardcoded paths are needed.

### Configuration

Add to your MCP client configuration (e.g., `mcp_config.json`):

```json
{
  "mcpServers": {
    "agent-ledger": {
      "command": "ledger-mcp"
    }
  }
}
```

This works when `ledger-mcp` is on your PATH. If it is not on PATH, use the absolute path to the binary:

```json
{
  "mcpServers": {
    "agent-ledger": {
      "command": "/Users/yourname/.local/bin/ledger-mcp"
    }
  }
}
```

### Antigravity IDE configuration

Place in `.agents/mcp_config.json` in your workspace root:

```json
{
  "mcpServers": {
    "agent-ledger": {
      "command": "ledger-mcp"
    }
  }
}
```

### Important

- The MCP client must launch `ledger-mcp` with the working directory set to your project (or any directory within it). The server walks upward to find the repository root.
- The target repository must have Agent Ledger initialized (`agent-ledger init`).
- Agent Ledger operates against the current Git workspace. Each repository has its own `.agent/` directory.

### MCP Tools

| Tool | Required Parameters | Description |
|---|---|---|
| `start_session` | — | Start a session (optional: `agent`, `model`) |
| `checkpoint` | — | Create an index-preserving Git checkpoint |
| `get_context` | — | Compile structured project context (optional: `task`) |
| `get_history` | — | Get session history (optional: `session_id`) |
| `get_handoff` | — | Get the latest handoff document |
| `explain_file` | `file_path` | Explain development history of a file |
| `record_decision` | `title`, `decision` | Record a decision (optional: `rationale`) |
| `record_discovery` | `title`, `finding` | Record a discovery |
| `record_failure` | `title`, `attempted`, `why` | Record a failure (optional: `lessons`) |
| `record_constraint` | `title`, `constraint` | Record a constraint (optional: `reason`) |
| `create_handoff` | `current_state`, `what_changed` | Create a handoff for the next session |
| `validate` | — | Validate ledger integrity |

### MCP Resources

Read-only context available to MCP clients:

| URI | Description |
|---|---|
| `agent://project/context` | Full compiled context |
| `agent://project/state` | Repository status and active session |
| `agent://project/architecture` | Package structure and relationships |
| `agent://project/decisions` | All recorded decisions |
| `agent://project/discoveries` | All recorded discoveries |
| `agent://project/failures` | All recorded failures |
| `agent://project/constraints` | All recorded constraints |
| `agent://session/current` | Active session details |
| `agent://session/history` | Session history |
| `agent://handoff/latest` | Most recent handoff |

---

## Agent-to-Agent Workflow

```
Agent A                          Agent B
  │                                │
  ├─ start session                 │
  ├─ write code, run tests         │
  ├─ checkpoint                    │
  ├─ record decisions/discoveries  │
  ├─ handoff                       │
  ├─ stop                          │
  │                                ├─ start session
  │                                ├─ get context / get handoff
  │                                ├─ continue development
  │                                ├─ checkpoint
  │                                └─ handoff → Agent C ...
```

### How context reconstruction works

When a fresh agent joins the repository:

1. `get_context` (or `agent-ledger context`) compiles:
   - **Git facts**: branch, HEAD, staged/unstaged/untracked files, package structure, test coverage
   - **Semantic knowledge**: decisions, discoveries, failures, constraints (filtered by recency)
   - **Latest handoff**: summary of the previous session's work and project state
2. `explain_file` provides file-specific development history from Git log and related decisions.

---

## Architecture

```
Git repository
  │
  ├── .agent/                 ← persistent human-readable ledger
  │   ├── project.md
  │   ├── state/current.json
  │   ├── sessions/<id>/
  │   │   ├── metadata.json
  │   │   ├── checkpoints.json
  │   │   └── handoff.md
  │   ├── decisions/
  │   ├── discoveries/
  │   ├── failures/
  │   └── constraints/
  │
  └── refs/agents/sessions/   ← Git checkpoint refs (not visible as files)
      └── <session-id>/
          ├── checkpoint-1
          └── checkpoint-2
```

### Components

| Layer | What it does |
|---|---|
| **Git** | Stores checkpoint commits as refs. Provides exact code history. |
| **`.agent/` directory** | Stores decisions, discoveries, failures, constraints, sessions, handoffs as human-readable Markdown and JSON. |
| **Context compiler** | Combines Git facts and semantic records into structured onboarding context. |
| **CLI (`agent-ledger`)** | Command-line interface for direct usage. |
| **MCP server (`ledger-mcp`)** | Stdio MCP interface so existing coding agents can use Agent Ledger. |

### The `.agent/` directory

The `.agent/` directory is part of your project's persistent development history. In user projects, it should normally be **committed to Git** so that knowledge persists across clones and contributors.

In the Agent Ledger source repository itself, `.agent/` is gitignored because it contains development state specific to working on Agent Ledger.

### Git-native checkpoint model

Checkpoints capture the complete working tree state (staged + unstaged + untracked) without disturbing the developer's Git workflow:

1. The current index is saved via `git write-tree`
2. All files are staged with `git add -A`
3. A tree object is written from the complete working tree
4. A checkpoint commit is created with `git commit-tree`
5. A ref is written under `refs/agents/sessions/<session-id>/checkpoint-<N>`
6. The original index and working tree are restored exactly

Checkpoints are inspectable with standard Git commands:

```bash
git show refs/agents/sessions/<session-id>/checkpoint-1
git show refs/agents/sessions/<session-id>/checkpoint-1:path/to/file.go
git diff refs/agents/sessions/<session-id>/checkpoint-1 refs/agents/sessions/<session-id>/checkpoint-2
```

---

## Releases

Releases are published as GitHub Releases at:

https://github.com/rahumanrahuu/agent-ledger/releases

### Release artifacts

Each release contains archives for all supported platforms:

```
agent-ledger_v0.2.1_darwin_arm64.tar.gz
agent-ledger_v0.2.1_darwin_amd64.tar.gz
agent-ledger_v0.2.1_linux_arm64.tar.gz
agent-ledger_v0.2.1_linux_amd64.tar.gz
agent-ledger_v0.2.1_windows_arm64.zip
agent-ledger_v0.2.1_windows_amd64.zip
checksums.txt
```

Each archive contains `agent-ledger` and `ledger-mcp` (with `.exe` on Windows).

### How releases are generated

Releases are built automatically by GitHub Actions when a version tag is pushed:

```bash
git tag v0.2.1
git push origin v0.2.1
```

The workflow runs `go test ./...`, then builds and packages binaries for all platforms using GoReleaser.

### Getting updates

Re-run the install command to upgrade to the latest release:

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/rahumanrahuu/agent-ledger/main/install.sh | sh

# Windows
irm https://raw.githubusercontent.com/rahumanrahuu/agent-ledger/main/install.ps1 | iex
```

---

## Project Structure

```
agent-ledger/
├── cmd/
│   ├── ledger/              # CLI entry point (installs as agent-ledger)
│   └── ledger-mcp/          # MCP server entry point
├── internal/
│   ├── checkpoint/          # Git ref checkpoint management
│   ├── context/             # Context compilation and formatting
│   ├── events/              # Semantic event creation and querying
│   ├── git/                 # Git CLI interface (plumbing commands)
│   ├── history/             # Session history aggregation
│   ├── repository/          # Git repository detection (upward walking)
│   ├── session/             # Session lifecycle management
│   └── storage/             # Local file storage (.agent/)
├── mcp/
│   ├── resources.go         # MCP resource handlers
│   └── tools.go             # MCP tool handlers
├── .github/workflows/       # GitHub Actions release workflow
├── .goreleaser.yaml          # GoReleaser cross-platform build config
├── install.sh               # Unix installer (macOS + Linux)
├── install.ps1              # Windows installer (PowerShell)
└── .agent/                  # Agent Ledger's own ledger (gitignored in this repo)
```

---

## Development Workflow

Agent Ledger follows a simple branch-based workflow:

1. **Development**: All normal development happens on the `dev` branch.
2. **Commit and Push**: Commit your changes and push them to `dev`.
3. **Merge**: Once tested and stable, merge the `dev` branch into the `main` branch.
4. **Release**: Create a version tag (e.g., `v0.3.0`) on the `main` branch.
5. **Publish**: The GitHub Actions release workflow will automatically build and publish the release. Tags created from `dev` or any other branch will fail the release verification.

---

## Running Tests

```bash
go test ./...
```

---

## Current Limitations

- **No cloud sync**: Local-only by design
- **Simple context compilation**: Rule-based keyword matching for task relevance, not AI-enhanced
- **Basic file explanation**: Uses Git history, not semantic analysis
- **Manual event recording**: Agents must explicitly call CLI/MCP tools to record knowledge
- **No multi-agent isolation**: Single active session at a time
- **No automatic session cleanup**: Stale active sessions are not auto-expired