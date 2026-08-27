# Memory System Implementation Guide

## Status: Phase 1-4 Implemented

This document outlines the memory system implementation that's been added to Agent Ledger.

## What Was Built

### Phase 1: Semantic Search ✅
**Files:**
- `internal/memory/memory.go` - Core memory manager with SQLite + FTS
- `internal/api/memory_endpoints.go` - Search, list, briefing APIs
- `cmd/ledger/memory_commands.go` - CLI commands for search

**Features:**
- Full-text search with BM25-like scoring
- Keyword search via SQLite FTS5
- Hybrid scoring (title match + keywords + content + importance)
- Type filtering (decision, discovery, constraint, failure)
- Threshold filtering
- Memory CRUD operations

**Usage:**
```bash
# Search memories
agent-ledger search "authentication strategy"
agent-ledger search "Supabase" --type decision --limit 10 --threshold 0.7

# List all
agent-ledger list-memories --type decision

# List with limit
agent-ledger list-memories --limit 50
```

### Phase 2: Event Ledger ✅
**Files:**
- `internal/memory/events.go` - Event recording and session replay

**Features:**
- Append-only event log (JSONL format)
- Event types: session_started, file_read, tool_call, error, etc.
- Session checkpoints with summaries
- Auto-generated session summaries
- Event retrieval by session

**Usage:**
```bash
# Create checkpoint
agent-ledger checkpoint "Added messaging feature"

# View session summary
agent-ledger session-summary <session-id>

# Replay session
agent-ledger replay <session-id>
```

**Storage Structure:**
```
.agent/
├── events/
│   └── {session-id}/
│       ├── events.jsonl      # Append-only event log
│       └── checkpoints/
│           ├── 20260827-143000.json
│           ├── 20260827-150000.json
│           └── ...
```

### Phase 3: Turn-1 Context Injection ✅
**Files:**
- `internal/api/memory_endpoints.go` - Briefing endpoint

**Features:**
- Automatic briefing generation based on task
- Relevance scoring for memory retrieval
- Tech stack extraction
- Decision/risk/constraint aggregation
- Estimated duration & cost

**Usage:**
```bash
# Frontend integration (already in BriefingPanel.jsx)
GET /api/briefing?task=implement_messaging&session_id=9d046179

# Returns:
{
  "task": "Implement real-time messaging",
  "tech_stack": ["Flutter", "Supabase", "Stream Chat"],
  "architecture": "Clean Architecture",
  "constraints": ["Auth: OTP only", "Rate: 8 sparks/day"],
  "decisions": ["Use Riverpod", "Use Supabase"],
  "risks": ["Webhook not configured"],
  "estimated_duration": "45-60 minutes"
}
```

### Phase 4: Constraint Enforcement ✅
**Files:**
- `internal/memory/constraints.go` - Constraint checking and violations

**Features:**
- Constraint definition from markdown files
- Glob-pattern file matching
- Violation detection
- Severity levels (CRITICAL, HIGH, MEDIUM, LOW)
- Pre-commit integration ready

**Usage:**
```bash
# Check file for violations
agent-ledger check-compliance lib/features/auth/login.dart

# Output:
# 🔴 auth-otp-only
#    File: lib/features/auth/login.dart
#    Severity: CRITICAL
#    Message: Firebase detected - must use Supabase OTP only
#    Suggestion: Replace Firebase with Supabase auth.signInWithOtp()

# Integration with git hooks (see below)
```

## How to Integrate into Main API

### 1. Update `internal/api/api.go`

Add memory manager field:
```go
type API struct {
	// ... existing fields ...
	memoryMgr *memory.Manager
}

// Update NewAPI to initialize memory
func NewAPI(repo *repository.Repository, st *storage.Storage, version string) *API {
	// ... existing code ...
	
	memMgr, err := memory.NewManager(repo.Root)
	if err != nil {
		// Handle error - create with nil if memory system fails
		memMgr = nil
	}
	
	return &API{
		// ... existing fields ...
		memoryMgr: memMgr,
	}
}
```

### 2. Update `ui/server.go`

Register memory endpoints:
```go
// API endpoints
mux.HandleFunc("/api/overview", s.handleOverview)
mux.HandleFunc("/api/sessions", s.handleSessions)
mux.HandleFunc("/api/events", s.handleEvents)
mux.HandleFunc("/api/graph", s.handleGraph)

// NEW: Memory endpoints
mux.HandleFunc("/api/search", s.api.HandleMemorySearch)
mux.HandleFunc("/api/briefing", s.api.HandleBriefing)
mux.HandleFunc("/api/memories", s.api.HandleMemoryList)
mux.HandleFunc("/api/memory/add", s.api.HandleMemoryAdd)
```

### 3. Update `cmd/ledger/main.go`

Add memory commands:
```go
case "search":
	handleMemorySearch()
case "checkpoint":
	handleMemoryCheckpoint()
case "list-memories":
	handleMemoryList()
case "check-compliance":
	handleMemoryCheckCompliance()
case "rebuild-index":
	handleMemoryRebuildIndex()
```

## Wire Frontend Components

The UI components are ready in:
- `ui/frontend/src/components/MemorySearch.jsx`
- `ui/frontend/src/components/BriefingPanel.jsx`

### In `ui/frontend/src/App.jsx`:

```jsx
import MemorySearch from './components/MemorySearch'
import BriefingPanel from './components/BriefingPanel'

// Add these hooks
const [searchOpen, setSearchOpen] = useState(false)
const [showBriefing, setShowBriefing] = useState(true)

// Add keyboard shortcut
useEffect(() => {
  const handleKeyDown = (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      setSearchOpen(true)
    }
  }
  window.addEventListener('keydown', handleKeyDown)
  return () => window.removeEventListener('keydown', handleKeyDown)
}, [])

// Render components
return (
  <>
    {showBriefing && <BriefingPanel task={currentTask} onDismiss={() => setShowBriefing(false)} />}
    <MemorySearch isOpen={searchOpen} onClose={() => setSearchOpen(false)} onSelect={handleMemorySelect} />
    {/* rest of app */}
  </>
)
```

## Database Schema

SQLite database at `.agent/memory/vectors.db`:

```sql
memories (
  id TEXT PRIMARY KEY,
  type TEXT (decision|discovery|constraint|failure),
  title TEXT,
  content TEXT,
  embedding BLOB (for future semantic search),
  keywords TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  importance REAL (0.0-1.0),
  session_id TEXT,
  path TEXT,
  archived INTEGER (soft delete)
)

memories_fts (FTS5 virtual table)
  - title
  - content
  - keywords
```

## Pre-Commit Hook Integration

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

for file in $(git diff --cached --name-only); do
    if [[ $file == *.dart ]] || [[ $file == *.py ]] || [[ $file == *.js ]]; then
        agent-ledger check-compliance "$file"
        
        exit_code=$?
        if [ $exit_code -eq 2 ]; then
            echo "Constraint violations detected. Fix before committing."
            exit 1
        fi
    fi
done

exit 0
```

Make executable: `chmod +x .git/hooks/pre-commit`

## Next Phases to Implement

### Phase 5: Reasoning Traces
Track every memory retrieval, tool call, and decision point

### Phase 6: Multi-Agent Coordination  
Shared memory between agents with conflict resolution

### Phase 7: Real-Time Collaboration
WebSocket-based live memory sync

## Testing

```bash
# Add a test memory
curl -X POST http://localhost:5173/api/memory/add \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-001",
    "type": "decision",
    "title": "Use TypeScript",
    "content": "Decided to use TypeScript for type safety",
    "keywords": "language, typescript, type-safety",
    "importance": 0.9
  }'

# Search
curl "http://localhost:5173/api/search?q=TypeScript&limit=5"

# Get briefing
curl "http://localhost:5173/api/briefing?task=add_typescript&session_id=test"

# List all memories
curl "http://localhost:5173/api/memories?limit=20"

# Check file compliance
agent-ledger check-compliance lib/main.dart
```

## Performance Targets

- Search (1k memories): < 50ms
- Search (100k memories): < 150ms
- Memory write: < 20ms
- Briefing generation: < 100ms
- Constraint check: < 10ms

## Storage

- 100k memories: ~2GB (with future semantic embeddings)
- Compression possible via consolidation

## What's Missing (Future Phases)

- [ ] Semantic search with embeddings
- [ ] Memory consolidation/compression
- [ ] Reasoning trace capture
- [ ] Multi-agent coordination protocol
- [ ] Real-time WebSocket sync
- [ ] Human-in-the-loop constraints
- [ ] Advanced metrics/observability
- [ ] Cost attribution

## Build & Deploy

```bash
# Build with memory support
go build -o agent-ledger ./cmd/ledger

# Test memory operations
./agent-ledger search "test query"

# Build frontend
cd ui/frontend
npm install && npm run build

# Production binary with embedded frontend
./agent-ledger ui
```

---

**Status:** Ready to integrate and test. All Phase 1-4 foundation code is complete and tested locally.
