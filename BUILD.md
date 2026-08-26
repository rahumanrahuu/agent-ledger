# Build Instructions for Agent Ledger UI

## Quick Build

### Development Build (Without Embedded Frontend)
```bash
go build -o bin/agent-ledger ./cmd/ledger
```

The UI will start on port 5173 and serve a development message (API endpoints still work).

### Production Build (With Embedded Frontend)

**On macOS/Linux:**
```bash
cd ui/frontend
npm install
npm run build
cd ../..
go build -o bin/agent-ledger ./cmd/ledger
```

**On Windows (PowerShell):**
```powershell
cd ui/frontend
npm install
npm run build
cd ../..
go build -o bin/agent-ledger.exe ./cmd/ledger
```

## What Happens in Each Build

### 1. Frontend Build
```bash
cd ui/frontend
npm run build
```

Creates optimized production assets in `ui/dist/`:
- JavaScript bundles (minified)
- CSS stylesheets (minified)
- index.html
- Other assets

### 2. Go Binary Build
```bash
go build ./cmd/ledger
```

The `ui/embed.go` file includes:
```go
//go:embed dist/*
var Dist embed.FS
```

This embeds all files from `ui/dist/` directly into the binary at compile time.

### 3. Runtime

When `agent-ledger ui` starts:
1. The server checks if embedded assets are available
2. If yes: serves from the embedded filesystem (no external dependencies)
3. If no: serves a development placeholder (development builds)

All API calls work identically in both cases.

## Docker Build Example

```dockerfile
FROM node:18 AS frontend-builder
WORKDIR /app/ui/frontend
COPY ui/frontend .
RUN npm install && npm run build

FROM golang:1.27
WORKDIR /app
COPY . .
COPY --from=frontend-builder /app/ui/dist ./ui/dist
RUN go build -o agent-ledger ./cmd/ledger

ENTRYPOINT ["./agent-ledger"]
```

## CI/CD Integration

### GitHub Actions Example
```yaml
- name: Build Frontend
  run: |
    cd ui/frontend
    npm install
    npm run build

- name: Build Binary
  run: go build -o agent-ledger ./cmd/ledger
```

### Verify Embedding

Check that your binary includes the frontend:
```bash
# On macOS/Linux
strings bin/agent-ledger | grep -i "index.html" | head -1

# On Windows PowerShell
Select-String "index.html" bin/agent-ledger.exe | Select-Object -First 1
```

If you see `index.html` in the output, the frontend is embedded.

## Troubleshooting

### "dist/*: no matching files found" error
- Frontend hasn't been built yet
- Run `cd ui/frontend && npm run build` first
- Ensure `ui/dist/` directory exists with built assets

### UI shows "development placeholder"
- Normal if frontend assets aren't embedded
- Build with the production build steps above to embed
- API endpoints still work in development mode

### Large Binary Size
- Expected: Vite frontend + React runtime + Go runtime = ~30-50MB
- Most of this is the Go runtime and vendor dependencies
- Use `go build -ldflags="-s -w"` to strip symbols (~10MB reduction)

## Development Workflow

For active frontend development:
```bash
# Terminal 1: Start Go backend
go run ./cmd/ledger ui --port 5173

# Terminal 2: Start Vite dev server (in ui/frontend)
npm run dev
```

Vite will proxy API calls to the Go server, providing hot module reloading for the frontend.
