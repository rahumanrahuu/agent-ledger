# Agent Ledger UI Setup & Usage

## Quick Start

### 1. Initialize Agent Ledger in a Git repository

```bash
cd your-repo
agent-ledger init
```

### 2. Start the UI server

```bash
agent-ledger ui [--port 5173]
```

The server will start and output:
```
Agent Ledger UI starting at http://localhost:5173
Press Ctrl+C to stop
```

Visit `http://localhost:5173` in your browser.

## Cross-Platform Build Instructions

### macOS & Linux

```bash
# 1. Build frontend assets
cd ui/frontend
npm install
npm run build

# 2. Build Go binary with embedded assets
cd ../..
go build -o bin/agent-ledger ./cmd/ledger

# 3. Run UI
./bin/agent-ledger ui
```

### Windows (PowerShell)

```powershell
# 1. Build frontend assets
cd ui/frontend
npm install
npm run build

# 2. Build Go binary with embedded assets
cd ..\..
go build -o bin/agent-ledger.exe .\cmd\ledger

# 3. Run UI
.\bin\agent-ledger.exe ui
```

## Development

### Building the Frontend

The UI is split into a Go backend and React+Vite frontend.

**First-time setup:**

```bash
cd ui/frontend
npm install
npm run build
```

This creates optimized assets in `ui/dist/` that will be embedded into the Go binary.

**During development** (if modifying frontend code):

```bash
cd ui/frontend
npm run dev
```

This starts a Vite dev server on port 5173 with hot module reloading. The dev proxy forwards `/api/*` requests to the Go backend.

### Embedding Frontend in Binary

Frontend embedding is handled by `ui/embed.go` using `go:embed`:

```go
package ui

import "embed"

//go:embed dist/*
var Dist embed.FS
```

When `npm run build` generates files in `ui/dist/`, running `go build ./cmd/ledger` automatically bundles all frontend assets directly into the standalone binary.

## Architecture

### Go Backend (`internal/api/api.go`)

Provides read-only REST API:
- `GET /api/overview` - Project metrics
- `GET /api/sessions` - List sessions
- `GET /api/sessions/:id` - Session details
- `GET /api/events?type=<filter>` - Event timeline
- `GET /api/graph` - Knowledge graph
- `GET /api/search?q=<query>` - Search results

### React Frontend (`ui/frontend/src/`)

Apple-designed web interface with:
- **Sidebar**: Navigation with `react-icons` (Overview, Sessions, Decisions, Discoveries, Checkpoints, Timeline)
- **Inspector**: Right-side detail panel
- **Views**: Overview, Sessions, Decisions, Discoveries, Checkpoints, Timeline
- **OS-Aware Shortcut**: Displays `⌘K` on macOS and `Ctrl+K` on Windows/Linux

## Design Principles

- **Apple Design Language**: SF Pro typography, restrained colors, subtle separators, macOS aesthetics
- **React Icons**: Modern Feather (`fi`) icons replacing plain unicode symbols
- **Local-first**: No external dependencies at runtime, all assets embedded in binary
- **Read-only**: All operations are informational only, no state mutations
- **Vanilla CSS**: CSS variables for theme system
- **Responsive**: Mobile, tablet, and desktop layouts

## Design System

CSS variables in `src/index.css`:

### Colors
- `--color-bg-primary`, `--color-bg-secondary`, `--color-bg-tertiary`
- `--color-text-primary`, `--color-text-secondary`, `--color-text-tertiary`
- `--color-blue`, `--color-green`, `--color-red`, `--color-yellow`

### Typography
- `--font-family-system`: -apple-system, BlinkMacSystemFont, etc.
- `--font-family-mono`: SF Mono, Monaco, etc.
- `--font-size-*`: xs through 5xl
- `--font-weight-*`: regular, medium, semibold, bold

### Spacing
- `--spacing-1` through `--spacing-12` (2px to 48px)
- Grid-based on 4px increment

### Shadows & Radius
- `--shadow-*`: sm, md, lg, xl
- `--radius-*`: sm, md, lg, xl, 2xl

## Implementation Status

### ✅ Complete
- Go backend REST API infrastructure
- React + Vite frontend scaffolding
- Apple design system CSS & React Icons integration
- Frontend asset embedding in binary (`ui/embed.go`)
- Cross-platform build support (macOS & Windows)
- OS-aware search shortcut (`⌘K` on Mac, `Ctrl+K` on Windows/Linux)
- Sidebar navigation & Inspector panel
- Overview, Sessions, Decisions, Discoveries, Checkpoints, and Timeline views
- Knowledge graph visualization null-safety checks

## Troubleshooting

**"Agent ledger not initialized"**: Run `agent-ledger init` first

**API returns empty data**: Ensure sessions exist: `agent-ledger sessions`

**Port already in use**: Use `--port` flag: `agent-ledger ui --port 5174`

**Blank screen / Assets not loading**: Ensure `npm run build` was executed inside `ui/frontend` prior to building the Go binary with `go build ./cmd/ledger`.
