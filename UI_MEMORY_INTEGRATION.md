# Agent Ledger UI: Memory System Integration Guide

## Overview
This document shows how to integrate the new memory features into the Agent Ledger web UI, enhancing the three-column Xcode-style layout with search, briefing, constraints, and reasoning traces.

## Components Added

### 1. MemorySearch Component
**File:** `ui/frontend/src/components/MemorySearch.jsx`  
**Purpose:** Full-text + semantic search with filters and real-time results

**Integration:**
```jsx
// In App.jsx or Sidebar.jsx
import MemorySearch from './components/MemorySearch'

// Add search trigger (e.g., Cmd+K shortcut)
const [searchOpen, setSearchOpen] = useState(false)

// Trigger on keyboard shortcut
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

// Render
<MemorySearch 
  isOpen={searchOpen} 
  onClose={() => setSearchOpen(false)}
  onSelect={(memory) => {
    setSelectedItem(memory)
    setSearchOpen(false)
  }}
/>
```

**Features:**
- Semantic + keyword hybrid search
- Type filtering (decisions, discoveries, constraints, failures)
- Time-range filtering
- Relevance threshold slider
- Search history
- Real-time results with 300ms debounce
- Scoring visualization

### 2. BriefingPanel Component
**File:** `ui/frontend/src/components/BriefingPanel.jsx`  
**Purpose:** Auto-generated context briefing shown on session/task start

**Integration:**
```jsx
// In App.jsx
import BriefingPanel from './components/BriefingPanel'

const [showBriefing, setShowBriefing] = useState(true)
const [currentTask, setCurrentTask] = useState('')

// Show briefing when new session starts
useEffect(() => {
  setShowBriefing(true)
  setCurrentTask('Your current task or objective')
}, [sessionId])

// Render in modal/panel
<BriefingPanel 
  task={currentTask}
  onDismiss={() => setShowBriefing(false)}
/>
```

**Features:**
- Task-based context retrieval
- Expandable sections (tech stack, decisions, risks, etc.)
- Constraint warnings (always visible)
- Estimated duration and token cost
- Recent decisions
- Known risks
- Next steps

### 3. ConstraintWarnings Component
**File:** `ui/frontend/src/components/ConstraintWarnings.jsx` (create)  
**Purpose:** Display active constraints and violations in real-time

```jsx
import { FiAlertTriangle, FiCheckCircle } from 'react-icons/fi'
import './ConstraintWarnings.css'

function ConstraintWarnings({ constraints, violations }) {
  return (
    <div className="constraints-panel">
      <h4>🔴 Active Constraints</h4>
      
      <div className="constraints-list">
        {constraints.map((c) => (
          <div key={c.id} className={`constraint ${c.violated ? 'violated' : 'active'}`}>
            {c.violated ? (
              <FiAlertTriangle className="violation-icon" />
            ) : (
              <FiCheckCircle className="active-icon" />
            )}
            
            <div className="constraint-info">
              <span className="constraint-name">{c.name}</span>
              <span className="constraint-description">{c.description}</span>
            </div>
            
            {c.violated && (
              <span className="violation-badge">VIOLATION</span>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

export default ConstraintWarnings
```

### 4. ReasoningTraces Component
**File:** `ui/frontend/src/components/ReasoningTraces.jsx` (create)  
**Purpose:** Show agent's reasoning process and memory retrievals

```jsx
import { FiCopy, FiTrendingUp, FiCheck, FiX } from 'react-icons/fi'
import './ReasoningTraces.css'

function ReasoningTraces({ traces, sessionId }) {
  const [expandedTrace, setExpandedTrace] = useState(null)
  
  if (!traces || traces.length === 0) return null
  
  return (
    <div className="reasoning-traces">
      <h4>🧠 Reasoning Trace</h4>
      <p className="trace-count">{traces.length} steps</p>
      
      <div className="traces-timeline">
        {traces.map((trace, i) => (
          <div 
            key={i} 
            className={`trace-item ${trace.type}`}
            onClick={() => setExpandedTrace(expandedTrace === i ? null : i)}
          >
            <span className="trace-icon">{getIcon(trace.type)}</span>
            <span className="trace-time">{trace.timestamp}</span>
            <span className="trace-label">{trace.label}</span>
            
            {trace.score && (
              <span className="trace-score">{(trace.score * 100).toFixed(0)}%</span>
            )}
            
            {expandedTrace === i && (
              <div className="trace-details">
                <pre>{JSON.stringify(trace.details, null, 2)}</pre>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
```

## Updated Layout

### Enhanced App.jsx Structure
```jsx
// Main three-column layout with memory features
<div className="app">
  <Sidebar 
    onSearch={() => setSearchOpen(true)}
    memories={memories}
    onMemorySelect={setSelectedItem}
  />
  
  <div className="center-column">
    {/* Briefing shown on top initially */}
    {showBriefing && <BriefingPanel task={task} />}
    
    {/* Main content area */}
    {renderView(currentView)}
    
    {/* Memory search modal */}
    <MemorySearch 
      isOpen={searchOpen}
      onClose={() => setSearchOpen(false)}
      onSelect={handleMemorySelect}
    />
  </div>
  
  <div className="right-column inspector">
    {/* Existing Inspector */}
    <Inspector item={selectedItem} />
    
    {/* NEW: Show active constraints in inspector */}
    {selectedItem && <ConstraintWarnings constraints={getRelevantConstraints(selectedItem)} />}
    
    {/* NEW: Show reasoning traces */}
    {session?.traces && <ReasoningTraces traces={session.traces} />}
  </div>
</div>
```

## Sidebar Enhancement

Add memory search trigger to sidebar:
```jsx
<div className="sidebar-footer">
  <button 
    className="memory-search-btn"
    onClick={onSearch}
    title="Search memories (Cmd+K)"
  >
    <FiSearch /> Search Memory
  </button>
</div>
```

## Keyboard Shortcuts

Add these shortcuts to enhance UX:
```jsx
// In App.jsx or a global hook
useEffect(() => {
  const handleKeyDown = (e) => {
    // Cmd/Ctrl+K: Search
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      setSearchOpen(true)
    }
    
    // Cmd/Ctrl+Shift+B: Show briefing
    if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'B') {
      e.preventDefault()
      setShowBriefing(true)
    }
    
    // Cmd/Ctrl+Shift+C: Show constraints
    if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'C') {
      e.preventDefault()
      setShowConstraints(!showConstraints)
    }
  }
  
  window.addEventListener('keydown', handleKeyDown)
  return () => window.removeEventListener('keydown', handleKeyDown)
}, [showConstraints])
```

## API Integration Points

### Search API
```
GET /api/search?q=query&type=decision&threshold=0.7&limit=10

Response:
{
  "results": [
    {
      "id": "9d046179-decision-001",
      "type": "decision",
      "title": "Use Supabase",
      "content": "...",
      "score": 0.92,
      "created_at": "2026-08-26T18:45:00Z",
      "session_id": "9d046179"
    }
  ]
}
```

### Briefing API
```
GET /api/briefing?task=implement_messaging&session_id=9d046179

Response:
{
  "task": "Implement real-time messaging",
  "tech_stack": ["Flutter", "Supabase", "Stream Chat"],
  "architecture": "Clean Architecture with core/data/features",
  "constraints": ["Auth: OTP only", "Rate: 8 sparks/day"],
  "decisions": ["Use Riverpod for state", "Supabase for backend"],
  "risks": ["Webhook not configured", "Real-time sync untested"],
  "next_steps": ["Set up webhooks", "Implement UI", "Test"],
  "estimated_duration": "45-60 minutes"
}
```

### Constraints API
```
GET /api/constraints?session_id=9d046179

Response:
{
  "constraints": [
    {
      "id": "auth-otp-only",
      "name": "OTP-Only Authentication",
      "severity": "CRITICAL",
      "description": "All auth must use Supabase OTP",
      "violated": false
    }
  ],
  "violations": []
}
```

### Reasoning Traces API
```
GET /api/session/9d046179/traces

Response:
{
  "traces": [
    {
      "type": "memory_retrieval",
      "timestamp": "14:35:12",
      "label": "Retrieved: Tech stack",
      "score": 0.94,
      "latency_ms": 87,
      "details": {...}
    }
  ]
}
```

## CSS Variables Used

The components use these CSS variables (update in App.css):
```css
:root {
  /* Colors */
  --color-blue: #0071e3;
  --color-green: #34c759;
  --color-yellow: #ff9500;
  --color-red: #ff3b30;
  --color-gray: #a0a0a0;
  
  /* Text */
  --color-text-primary: #000;
  --color-text-secondary: #666;
  --color-text-tertiary: #999;
  
  /* Background */
  --color-bg-primary: #fff;
  --color-bg-secondary: #f5f5f7;
  --color-bg-tertiary: #efefef;
  --color-border: #e5e5e7;
  
  /* Shadows */
  --shadow-md: 0 1px 3px rgba(0,0,0,0.1);
  --shadow-lg: 0 2px 8px rgba(0,0,0,0.12);
  
  /* Spacing & Typography */
  --font-size-xs: 12px;
  --font-size-sm: 14px;
  --font-size-md: 16px;
  --font-size-lg: 18px;
  --font-size-xl: 20px;
  
  --spacing-1: 4px;
  --spacing-2: 8px;
  --spacing-3: 12px;
  --spacing-4: 16px;
  --spacing-5: 20px;
  --spacing-6: 24px;
  
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;
}
```

## Step-by-Step Implementation

### Week 1: Core Search
1. Create `MemorySearch.jsx` and CSS
2. Wire up `/api/search` endpoint
3. Add keyboard shortcut (Cmd+K)
4. Test with real memories

### Week 2: Briefing
1. Create `BriefingPanel.jsx` and CSS
2. Wire up `/api/briefing` endpoint
3. Show on session start
4. Test context accuracy

### Week 3: Constraints & Traces
1. Create `ConstraintWarnings.jsx`
2. Create `ReasoningTraces.jsx`
3. Wire up corresponding APIs
4. Display in inspector panel

### Week 4: Polish
1. Add animations
2. Add keyboard shortcuts
3. Optimize API calls
4. User testing

## Performance Considerations

- **Search debounce:** 300ms (prevent excessive API calls)
- **Briefing cache:** Cache for current session (1 hour TTL)
- **Traces pagination:** Load first 50, lazy-load on scroll
- **CSS-in-JS:** Minimize re-renders with React.memo

## Accessibility

- All interactive elements keyboard navigable
- ARIA labels for screen readers
- Color not sole indicator (use icons + text)
- High contrast mode support

## Testing Checklist

- [ ] Search works with various queries
- [ ] Briefing shows on new session
- [ ] Constraints update in real-time
- [ ] Reasoning traces capture all steps
- [ ] Mobile responsive
- [ ] Keyboard shortcuts work
- [ ] API error handling
- [ ] Performance metrics acceptable
- [ ] Accessibility passes WCAG 2.1

---

**Next:** Implement these components one phase at a time, starting with MemorySearch (Phase 1), then BriefingPanel (Phase 3), then Constraints & Traces (Phases 4-5).
