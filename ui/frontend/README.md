# Agent Ledger UI Frontend

React + Vite frontend for the Agent Ledger web UI, styled with Apple design language principles.

## Development

```bash
cd ui/frontend
npm install
npm run dev
```

The dev server runs on `http://localhost:5173` with proxying to the Go API at `http://localhost:5173/api`.

## Building

```bash
npm run build
```

Builds production assets to `../dist/`. These will be embedded in the Go binary via `go:embed`.

## Architecture

- **Design System**: Apple-inspired with CSS variables for light/dark mode
- **Component Structure**:
  - `Sidebar`: Navigation between views
  - `Inspector`: Right-side detail panel
  - `Overview`: Project metrics and knowledge graph
  - `Sessions`, `Decisions`, `Discoveries`, `Checkpoints`, `Timeline`: Entity views

## Styling

All styling uses Vanilla CSS with CSS variables (no Tailwind). Colors, typography, spacing, and other design tokens are defined in `src/index.css`.

## API

The frontend expects the following API endpoints from the Go backend:

- `GET /api/overview` - Project overview with metrics
- `GET /api/sessions` - List of all sessions
- `GET /api/sessions/:id` - Details of a specific session
- `GET /api/events?type=<filter>` - Timeline events
- `GET /api/graph` - Knowledge graph nodes and edges
- `GET /api/search?q=<query>` - Search results
