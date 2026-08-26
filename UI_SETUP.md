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

## Development

### Building the Frontend

The UI is split into Go backend and React+Vite frontend.

**First-time setup:**

```bash
cd ui/frontend
npm install
npm run build
```

This creates optimized assets in `ui/dist/` that will be embedded in the binary.

**During development** (if you want to modify frontend code):

```bash
cd ui/frontend
npm run dev
```

This starts a development server on port 5173 with hot module reloading. The proxy automatically forwards `/api/*` requests to the Go server.

### Embedding Frontend in Binary

Once the frontend is built, it can be embedded in the Go binary using `go:embed`:

1. Build the frontend: `npm run build` in `ui/frontend`
2. Update `ui/server.go` to add:
   ```go
   //go:embed dist/*
   var UI embed.FS
   ```
3. Rebuild Go binary: `go build ./cmd/ledger`

For now, the server runs in development mode and serves a placeholder UI until the frontend assets are embedded.

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
- **Sidebar**: Navigation between views
- **Inspector**: Right-side detail panel
- **Views**: Overview, Sessions, Decisions, Discoveries, Checkpoints, Timeline

## Design Principles

- **Apple Design Language**: SF Pro typography, restrained colors, subtle separators, macOS aesthetics
- **Local-first**: No external dependencies, all assets embedded in binary
- **Read-only**: All operations are informational only, no state mutations
- **Vanilla CSS**: No Tailwind; CSS variables for theme system
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

## Light/Dark Mode

Automatically follows system preference via `prefers-color-scheme` media query. Users can override with explicit theme choice (implemented via `data-theme` attribute).

## Performance

- **Minimal JS**: Only React + Vite runtime
- **Offline**: No external API calls
- **Fast**: Embedded assets, local only
- **Accessible**: Semantic HTML, keyboard navigation, focus management

## Implementation Status

### ✅ Complete
- Go backend infrastructure
- REST API endpoints (basic)
- React + Vite scaffolding
- Apple design system CSS
- Sidebar navigation
- Inspector panel
- Overview view (with metrics cards)
- Sessions view (displays sessions from API)
- Stub views for other entity types

### 🚧 In Progress
- Full API implementation (reading from .agent directory)
- Markdown rendering for decisions/discoveries
- Knowledge graph visualization (custom SVG)
- Timeline view
- Search functionality
- Dark mode testing

### 📋 Todo
- Keyboard navigation polish
- Accessibility audit
- Mobile testing
- Frontend asset embedding in binary
- Production build & release process

## Troubleshooting

**"Agent ledger not initialized"**: Run `agent-ledger init` first

**API returns empty data**: Ensure sessions exist: `agent-ledger sessions`

**Port already in use**: Use `--port` flag: `agent-ledger ui --port 5174`

**Frontend shows placeholder**: Frontend assets are not yet embedded. This is normal in development. Build the frontend with `npm run build` to embed assets.
