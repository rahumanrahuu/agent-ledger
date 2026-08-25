# Agent Ledger

A local-first, Git-native system that preserves the development history and context of AI coding-agent sessions so that different agents can continue work without losing previous decisions, discoveries, failures, constraints, and relevant code history.

---

## Getting Started

### 1. Prerequisites

- **Go**: Go 1.21 or newer (project uses Go 1.27 toolchain directives in `go.mod`)
- **Git**: Git CLI installed and available in your `PATH`
- **Target Repository**: Target workspace must be an initialized Git repository

### 2. Clone the Repository

```bash
git clone https://github.com/yourusername/agent-ledger.git
cd agent-ledger
```

### 3. Build the Binaries

Build the CLI (`ledger`) and the MCP server (`ledger-mcp`):

```bash
go build -o ledger ./cmd/ledger
go build -o ledger-mcp ./cmd/ledger-mcp
```

*(Optional)* Move the binaries into your `PATH` (for example, `/usr/local/bin`):

```bash
cp ledger /usr/local/bin/agent-ledger
cp ledger-mcp /usr/local/bin/ledger-mcp
```

### 4. Initialize Agent Ledger

Agent Ledger requires the target workspace to be an existing Git repository. Inside your repository root, run:

```bash
agent-ledger init
# or: ./ledger init
```

This creates the `.agent/` directory structure and records initial repository metadata in `.agent/project.md`.

### 5. Start a Session

Start a session before performing agent tasks. The `--agent` and `--model` flags accept arbitrary, optional strings:

```bash
agent-ledger start --agent claude-code --model claude-3-5-sonnet
# or start with arbitrary identifiers or no flags:
agent-ledger start --agent cursor
agent-ledger start
```

---

## Basic CLI Workflow

The CLI supports the complete session and knowledge management lifecycle:

```bash
# 1. Start a session
agent-ledger start --agent devin --model swe-1.6

# 2. Check repository status and active session
agent-ledger status

# 3. Retrieve compiled onboarding context (optionally targeted to a task)
agent-ledger context
agent-ledger context --task "implement repository unit tests"

# 4. Create an index-preserving checkpoint of current working tree
agent-ledger checkpoint

# 5. Record semantic knowledge during development
agent-ledger decision --title "Use table-driven tests" --decision "Use table-driven tests for Git operations" --rationale "Systematically tests multiple command exit codes"
agent-ledger discovery --title "macOS temp path formatting" --finding "Temp directories on macOS resolve with /private prefix"
agent-ledger failure --title "Direct branch checkout" --attempted "Checked out branch directly in test" --why "Polluted developer workspace" --lessons "Use t.TempDir() isolated repos"
agent-ledger constraint --title "No cloud dependencies" --constraint "All state must remain strictly local in Git/.agent" --reason "Offline capability and data privacy"

# 6. Explain file-specific development history
agent-ledger explain internal/repository/repository.go

# 7. Create a handoff for the next session
agent-ledger handoff --state "Repository tests implemented" --changed "Added repository_test.go with full test coverage"

# 8. Stop the active session
agent-ledger stop

# 9. View session history and validate ledger integrity
agent-ledger history
agent-ledger sessions
agent-ledger validate
```

---

## Agent-to-Agent Handoff Workflow

When multiple agents work sequentially on the same repository, Agent Ledger bridges context gaps:

```
Agent A
  ↓
start session (agent-ledger start / start_session)
  ↓
work (code changes, tests)
  ↓
checkpoint (agent-ledger checkpoint / checkpoint)
  ↓
record important knowledge (decision, discovery, failure, constraint)
  ↓
handoff (agent-ledger handoff / create_handoff)
  ↓
Agent B
  ↓
start session (agent-ledger start / start_session)
  ↓
get context (agent-ledger context / get_context & get_handoff)
  ↓
continue development
```

### Reconstructing Context in a Fresh Session

A fresh agent joining the repository reconstructs context by:
1. Starting or attaching to a session (`start_session` or `agent-ledger start`).
2. Fetching compiled context (`get_context` or `agent-ledger context`), which combines:
   - **Objective Git Facts**: Current branch, HEAD commit, unstaged/staged/untracked file status, package architecture overview, and test coverage status per package.
   - **Semantic Knowledge**: Prior decisions, discoveries, known failures, and constraints filtered by recency and relevance.
   - **Latest Handoff**: Summary of recent work and current project state.
3. Querying specific file histories with `explain_file` when modifying existing modules.

---

## MCP Server Setup

Agent Ledger includes a Model Context Protocol (MCP) server executable (`ledger-mcp`) communicating over stdio.

### MCP Configuration (`mcp_config.json`)

Add `agent-ledger` to your client configuration using the absolute path to `ledger-mcp`:

```json
{
  "mcpServers": {
    "agent-ledger": {
      "command": "/path/to/agent-ledger/ledger-mcp"
    }
  }
}
```

> **Note**: The MCP server requires its working directory to be within an initialized Git repository containing `.agent/`. Ensure the MCP client launches from or sets the working directory to the target repository root.

### Available MCP Tools

The `ledger-mcp` binary provides 12 MCP tools:

| Tool | Parameters | Description |
|---|---|---|
| `start_session` | `agent` (string, optional), `model` (string, optional) | Starts a new session and records session metadata. |
| `checkpoint` | *none* | Creates an index-preserving Git checkpoint ref. |
| `get_context` | `task` (string, optional) | Compiles structured project context and semantic knowledge. |
| `get_history` | `session_id` (string, optional) | Returns session history across past sessions. |
| `get_handoff` | *none* | Returns the latest session handoff. |
| `explain_file` | `file_path` (string, required) | Explains the development history and changes for a specific file. |
| `record_decision` | `title` (string, required), `decision` (string, required), `rationale` (string, optional) | Records an architectural or implementation decision. |
| `record_discovery` | `title` (string, required), `finding` (string, required) | Records a technical or codebase discovery. |
| `record_failure` | `title` (string, required), `attempted` (string, required), `why` (string, required), `lessons` (string, optional) | Records a failed approach and lessons learned. |
| `record_constraint` | `title` (string, required), `constraint` (string, required), `reason` (string, optional) | Records an operational or architectural constraint. |
| `create_handoff` | `current_state` (string, required), `what_changed` (string, required) | Creates a handoff document for subsequent sessions. |
| `validate` | *none* | Validates ledger consistency, checkpoint refs, and files. |

### Available MCP Resources

The MCP server also exposes read-only resource URIs:

- `agent://project/context`: Full compiled context document.
- `agent://project/state`: Current repository dirty/staged status and active session info.
- `agent://project/architecture`: Package structure and relationships.
- `agent://project/decisions`: All recorded decisions.
- `agent://project/discoveries`: All recorded discoveries.
- `agent://project/failures`: All recorded failures and lessons.
- `agent://project/constraints`: All recorded constraints.
- `agent://session/current`: Active session details.
- `agent://session/history`: Comprehensive session history.
- `agent://handoff/latest`: Most recent handoff text.

---

## Storage & The `.agent/` Directory

All machine state and semantic files reside in the `.agent/` directory in the repository root:

```
.agent/
├── project.md               # Repository root, branch, HEAD, and init timestamp
├── state/
│   └── current.json         # Active session ID and runtime status
├── sessions/
│   └── <session-id>/
│       ├── metadata.json    # Session ID, agent, model, branch, timestamps
│       ├── checkpoints.json # Checkpoints created during this session
│       └── handoff.md       # Session handoff notes
├── decisions/
│   └── <timestamp>-<id>.md  # Markdown records of decisions and rationales
├── discoveries/
│   └── <timestamp>-<id>.md  # Markdown records of technical discoveries
├── failures/
│   └── <timestamp>-<id>.md  # Markdown records of failed attempts & lessons
└── constraints/
    └── <timestamp>-<id>.md  # Markdown records of project constraints
```

---

## Git-Native Checkpoint Model

Agent Ledger uses Git's internal object database to create lightweight snapshots without interfering with the developer's working directory or standard branch commits:

1. **Dedicated Ref Namespace**: Checkpoints are written as Git commits under `refs/agents/sessions/<session-id>/checkpoint-<N>`. They do not create branches or alter `HEAD`.
2. **Index-Preserving**: The checkpoint algorithm stages and commits working tree changes via Git low-level plumbing (`git write-tree`, `git commit-tree`, `git update-ref`), then restores the exact prior index and working tree state. Any staged, unstaged, or untracked state remains untouched.
3. **Inspectable with Native Git**: Any checkpoint commit can be directly examined using standard Git commands:
   ```bash
   # View checkpoint details
   git show refs/agents/sessions/<session-id>/checkpoint-1

   # Inspect file at a checkpoint
   git show refs/agents/sessions/<session-id>/checkpoint-1:path/to/file.go

   # Compare two checkpoints
   git diff refs/agents/sessions/<session-id>/checkpoint-1 refs/agents/sessions/<session-id>/checkpoint-2
   ```

---

## Project Structure

```
agent-ledger/
├── cmd/
│   ├── ledger/           # CLI entry point
│   └── ledger-mcp/       # Stdio MCP server entry point
├── internal/
│   ├── checkpoint/       # Git ref checkpoint management
│   ├── context/          # Context compilation and relevance ranking
│   ├── events/           # Semantic event creation (decisions, discoveries, etc.)
│   ├── git/              # Git CLI interface and plumbing commands
│   ├── history/          # Session history aggregation
│   ├── repository/       # Git repository detection and status parsing
│   ├── session/          # Session lifecycle management
│   └── storage/          # Local file storage layer (.agent/)
└── mcp/
    ├── resources.go      # MCP resource handlers
    └── tools.go          # MCP tool handlers
```

---

## Running Tests

Run the full Go test suite across all internal packages and the MCP server:

```bash
go test ./...
```