# Agent Ledger — Complete UI/UX + API Integration

You are working on the **Agent Ledger** repository.

The current application has a partially implemented React UI and Go backend/API. The UI is currently inconsistent, some pages are incomplete, some APIs are not connected correctly, and some screens show incorrect/empty data even though the backend contains the data.

Your job is to **fully audit, fix, and complete the application**.

## IMPORTANT

Do NOT blindly rewrite the project.

First understand the existing architecture, backend APIs, data models, MCP server, storage layer, WebSocket implementation, and current frontend.

The goal is:

> **A production-quality Agent Ledger web application where every existing backend capability has an appropriate UI and every UI action actually works against the real backend.**

Do not replace real data with mock data.

Do not hardcode counts.

Do not create fake API responses.

Do not remove working backend functionality just to make the UI easier.

---

# 1. FIRST: AUDIT THE ENTIRE REPOSITORY

Before changing code, inspect the complete repository.

Understand:

* Go backend
* MCP server
* internal API
* memory system
* session system
* timeline/events
* decisions
* discoveries
* failures
* constraints
* checkpoints
* handoffs
* Git integration
* knowledge graph
* WebSocket/live updates
* React frontend
* routing
* components
* API client
* state management
* TypeScript types
* existing CSS/design system

Pay particular attention to:

```text
cmd/
mcp/
internal/
ui/
```

and all frontend source directories.

Find the actual implementations for:

```text
/api/overview
/api/sessions
/api/sessions/{id}
/api/events
/api/graph
/api/search
/api/memories
/api/live
```

Also inspect whether additional API endpoints already exist but are not represented in the UI.

Create a mental map of:

```text
Backend API
     ↓
API client
     ↓
React state
     ↓
Page/component
     ↓
User interaction
```

Identify every broken link in that chain.

---

# 2. DO NOT CHANGE THE BACKEND UNLESS NECESSARY

The backend already contains important functionality.

Prefer fixing the frontend/API integration instead of replacing backend implementations.

Only modify backend code when:

1. An API is actually broken.
2. Required data is unavailable.
3. The frontend cannot correctly consume an existing API because of a backend bug.
4. A small missing endpoint is genuinely required for an existing feature.

If you modify backend code, preserve compatibility with the MCP server and existing storage.

Do not break:

* MCP tools
* Git-native storage
* SQLite
* FTS5
* memory persistence
* WebSocket updates
* session persistence

---

# 3. USE THE LAST UPLOADED IMAGE AS THE PRIMARY UI INSPIRATION

Use the **last uploaded SaaS Startup UI screenshot** as the visual inspiration.

Do NOT copy the branding/content.

Use its design language.

The target should feel like a modern premium SaaS developer/productivity application.

### Visual characteristics to adopt

* clean white/light interface
* generous whitespace
* restrained borders
* subtle shadows
* rounded cards
* clear hierarchy
* professional typography
* compact but readable navigation
* polished dashboard layouts
* consistent spacing
* clear primary/secondary actions
* subtle accent color
* minimal visual noise
* strong alignment
* consistent card dimensions
* modern SaaS settings/dashboard aesthetic

The screenshot is inspiration for **layout, spacing, hierarchy, polish and UX**, not something to reproduce literally.

Agent Ledger should retain its own identity.

---

# 4. CREATE ONE CONSISTENT DESIGN SYSTEM

The current application feels like several different UIs were built separately.

Fix that.

Create shared design primitives/components for:

* AppShell
* Sidebar
* TopBar
* PageHeader
* Breadcrumbs
* Cards
* StatCards
* Buttons
* IconButtons
* Badges
* StatusIndicators
* Tabs
* Filters
* SearchInput
* Dropdowns
* EmptyStates
* LoadingStates
* ErrorStates
* Skeletons
* Tables
* Lists
* TimelineItems
* InspectorPanel
* Modal/Dialog
* Toasts
* Tooltips
* Pagination where necessary

Every page must use the same design language.

Do not repeatedly implement slightly different cards, buttons or spacing.

---

# 5. APPLICATION SHELL

Build a polished application shell.

### Sidebar

The sidebar should contain:

**Agent Ledger**

Repository/project identity

Navigation:

* Overview
* Memories
* Timeline
* Sessions
* Knowledge Graph

If the backend supports additional major concepts that deserve dedicated screens, add them appropriately.

The active navigation item should be obvious but subtle.

Sidebar should remain stable while navigating.

### Top navigation

Include:

* project/repository context
* connection/live status
* useful global actions
* search where appropriate

Avoid unnecessary visual clutter.

---

# 6. OVERVIEW PAGE

The Overview page should become the real project dashboard.

Use actual `/api/overview` data.

Include:

### Project header

* Project name
* Repository path
* Branch
* Current commit
* Version/state
* Last activity
* Live connection status

### Metrics

Display real values for:

* Sessions
* Decisions
* Discoveries
* Failures
* Constraints
* Checkpoints
* Memories

Do not hardcode these.

Cards should be clickable and navigate to the relevant page.

For example:

```text
Sessions → Sessions
Decisions → Timeline filtered to decisions
Discoveries → Timeline filtered to discoveries
Checkpoints → relevant checkpoint/session information
Memories → Memories
```

### Recent activity

Show meaningful recent events using real backend data.

### Knowledge graph preview

Show a useful preview of the actual graph.

Provide:

> Open knowledge graph →

Do not create a decorative fake graph.

---

# 7. MEMORIES PAGE — IMPORTANT

The current Memories page is broken.

It currently shows:

```text
0 memories
No memories found
```

even though the project clearly contains decisions/discoveries/etc.

Investigate why.

Trace:

```text
/api/memories
      ↓
frontend API request
      ↓
response
      ↓
memory transformation
      ↓
rendering
```

Fix the actual problem.

Do NOT simply change the UI to display fake numbers.

The Memories page should support:

* real memory listing
* search
* type filtering
* relevance where available
* importance
* timestamps
* session association
* memory type
* details/inspector
* empty state only when genuinely empty
* loading state
* API error state

Memory types should include whatever the backend actually supports, such as:

* Decision
* Discovery
* Failure
* Constraint

If deletion/archive functionality is supported by the backend, expose it safely in the UI.

---

# 8. TIMELINE PAGE

The Timeline page should be a polished chronological activity feed.

Use real `/api/events`.

Support filters:

* All
* Decisions
* Discoveries
* Failures
* Constraints
* Checkpoints

Each event should show:

* type
* title
* timestamp
* short summary
* relevant session
* useful metadata

Clicking an event should open a proper detail/inspector view.

The inspector should feel like a native part of the application, not a random floating box.

Support deep linking where practical.

---

# 9. SESSIONS PAGE

Use real `/api/sessions`.

Display sessions as professional cards/list rows.

Each session should show useful information such as:

* session name
* agent
* model
* branch
* start time
* end time/status
* event count
* associated decisions/discoveries/checkpoints where available

Clicking a session should open a detailed session page or inspector.

The session detail should provide:

* session metadata
* timeline
* decisions
* discoveries
* failures
* constraints
* checkpoints
* handoff information
* relevant memories

Use the actual API.

Do not hardcode:

```text
claude-fable-5
ACTIVE
```

etc.

---

# 10. KNOWLEDGE GRAPH

The graph is one of the most important parts of Agent Ledger.

Make it feel like a real developer knowledge graph.

Use the real:

```text
/api/graph
```

data.

The graph must support:

* pan
* zoom
* node selection
* edge relationships
* node type styling
* search
* filtering
* auto layout
* fit-to-screen
* reset view
* selected-node inspector

Node types should visually distinguish:

* Session
* Decision
* Discovery
* Failure
* Constraint
* Checkpoint
* other real backend types

Do not use fake relationships.

Do not randomly generate graph connections.

The graph should represent the actual Agent Ledger relationships.

### Search

Searching should actually filter/find nodes.

### Filter

Filtering by node type should actually work.

### Inspector

When a node is selected, show its real data.

For example:

```text
TYPE
Decision

TITLE
Use Fable model for this session

ID
...

TIMESTAMP
...

SESSION
...

CONTENT
...

RELATED NODES
...
```

If a node has related events, make them clickable.

---

# 11. GLOBAL SEARCH

If `/api/search` supports cross-entity searching, build a proper global search experience.

Search across:

* sessions
* decisions
* discoveries
* failures
* constraints
* checkpoints
* memories

Results should clearly show:

```text
Type
Title
Relevant snippet
Timestamp
Session
```

Clicking a result should navigate to the appropriate entity.

---

# 12. REAL-TIME UPDATES

The application already has:

```text
/api/live
```

WebSocket functionality.

Make the frontend properly consume it.

When new ledger activity arrives:

* update relevant data
* update timeline
* update metrics
* update sessions
* update graph where appropriate
* show a subtle live indicator
* avoid full-page reloads

Connection states should be:

```text
Connected
Connecting
Reconnecting
Disconnected
```

Do not show "Reconnecting..." permanently when the socket is actually connected.

Investigate the existing WebSocket implementation instead of replacing it.

---

# 13. LOADING / ERROR / EMPTY STATES

Every API-driven page must have three deliberate states.

### Loading

Use skeletons rather than a blank white screen.

### Error

Show:

```text
Something went wrong

Unable to load this data.

Retry
```

Include useful technical information where appropriate.

### Empty

Only show empty states when the backend genuinely returns no records.

Example:

```text
No memories yet

Agent decisions, discoveries and other durable knowledge
will appear here as they are recorded.
```

Do not confuse API failure with an empty database.

---

# 14. INSPECTOR / DETAIL UX

The current inspector panels feel inconsistent.

Create one reusable inspector component.

It should:

* slide in smoothly
* have clear hierarchy
* have close button
* support scrolling
* show structured metadata
* display full content
* provide navigation to related records
* work consistently across sessions, timeline events, graph nodes and memories

On smaller screens it should become a full-screen detail view.

---

# 15. RESPONSIVE DESIGN

The application must work at:

* 1440px+
* 1280px
* 1024px
* tablet
* mobile

Do not let:

* graph overflow the page
* inspector overlap content incorrectly
* cards become unreadable
* sidebar destroy usable width
* tables overflow uncontrollably

The desktop experience is the priority, but responsive behavior must be intentional.

---

# 16. VISUAL QUALITY BAR

Do NOT produce a generic AI-generated dashboard.

Avoid:

* excessive gradients
* giant rounded cards
* excessive shadows
* random colors
* oversized typography
* inconsistent icons
* excessive animations
* unnecessary glassmorphism
* decorative elements with no purpose
* huge empty spaces caused by broken layouts
* placeholder content
* fake statistics

Aim for:

**Linear / Vercel / modern developer SaaS quality**

but keep Agent Ledger's own visual identity.

---

# 17. TYPOGRAPHY AND SPACING

Establish consistent values.

Use a restrained type scale.

Example:

```text
Page title
20–28px

Section title
16–18px

Body
13–15px

Metadata
11–13px
```

Use consistent spacing throughout the app.

Avoid every component inventing its own padding.

---

# 18. ICONS

Use one consistent icon library already present in the project.

Do not mix random icon styles.

Icons should communicate meaning rather than decorate every line.

---

# 19. API CLIENT

Audit the frontend API layer.

Create a clean typed API client if one does not already exist.

For example:

```text
getOverview()
getSessions()
getSession(id)
getEvents()
getGraph()
search()
getMemories()
```

Centralize:

* request handling
* error handling
* JSON parsing
* WebSocket handling
* API base URL
* types

Do not scatter raw `fetch()` calls throughout components unless there is a strong reason.

---

# 20. TYPE SAFETY

Use the actual backend response structures.

Do not make everything:

```ts
any
```

Create proper types/interfaces for:

```text
Session
Decision
Discovery
Failure
Constraint
Checkpoint
Memory
GraphNode
GraphEdge
Event
Overview
SearchResult
```

Adapt these to the real backend structures after inspecting them.

---

# 21. ROUTING

Every major page should have a proper route.

At minimum:

```text
/overview
/memories
/timeline
/sessions
/knowledge-graph
```

If useful:

```text
/sessions/:id
/memories/:id
/events/:id
```

Use the routing system already present in the project.

Avoid state-only navigation that breaks browser back/forward behavior.

---

# 22. FIX CURRENT VISUAL PROBLEMS

Specifically inspect the screenshots/current implementation for problems such as:

* inconsistent content width
* excessive blank space
* graph occupying the wrong proportions
* inspector appearing detached from the application
* tiny unreadable graph labels
* weak hierarchy
* inconsistent card styling
* poor responsive behavior
* Memories showing incorrect zero state
* disconnected/reconnecting status problems
* pages feeling like separate applications
* insufficient API data being surfaced
* missing actions
* inconsistent loading behavior

Fix the underlying implementation rather than hiding symptoms.

---

# 23. DO NOT REMOVE EXISTING FUNCTIONALITY

Before deleting any component, endpoint, API function, hook, or backend package, determine whether something depends on it.

Preserve existing functionality.

Improve it rather than replacing it unnecessarily.

---

# 24. TEST THE FULL APPLICATION

After implementation:

### Backend

Run the existing Go tests.

### Frontend

Run:

```bash
npm run build
```

or the project's equivalent.

Also run the development server.

### API testing

Verify every endpoint manually.

Check:

```text
/api/overview
/api/sessions
/api/sessions/{id}
/api/events
/api/graph
/api/search
/api/memories
/api/live
```

### UI testing

Actually navigate through:

```text
Overview
↓
Memories
↓
Timeline
↓
Sessions
↓
Knowledge Graph
```

Test:

* search
* filters
* clicking cards
* inspectors
* graph interactions
* session details
* WebSocket updates
* browser refresh
* browser back/forward
* empty state
* loading state
* API failure state

---

# 25. IMPORTANT DEBUGGING RULE

If something displays:

```text
0
No data
No memories
Disconnected
```

do NOT immediately modify the UI.

Trace the data.

For every suspicious value:

```text
UI value
↓
React state
↓
API response
↓
HTTP request
↓
Go handler
↓
service
↓
storage
```

Find where the value becomes incorrect.

Fix the actual source.

---

# 26. FINAL PRODUCT EXPERIENCE

When finished, Agent Ledger should feel like a coherent developer product.

The user should immediately understand:

```text
What project am I looking at?
What has happened?
What decisions were made?
What knowledge has been discovered?
What sessions occurred?
What changed recently?
How are these things related?
What should the next AI agent know?
```

The UI should make the answer to those questions obvious.

---

# 27. IMPLEMENTATION ORDER

Follow this order:

### Phase 1

Audit repository and APIs.

### Phase 2

Fix API client and data types.

### Phase 3

Create shared design system/application shell.

### Phase 4

Fix Overview.

### Phase 5

Fix Memories and verify `/api/memories`.

### Phase 6

Fix Timeline.

### Phase 7

Fix Sessions and session details.

### Phase 8

Rebuild/polish Knowledge Graph using real `/api/graph`.

### Phase 9

Implement global search.

### Phase 10

Integrate WebSocket live updates.

### Phase 11

Responsive behavior.

### Phase 12

Loading/error/empty states.

### Phase 13

Full testing and bug fixing.

---

# 28. VERY IMPORTANT — DO NOT STOP AFTER MAKING IT LOOK GOOD

A visually beautiful frontend that does not work is NOT acceptable.

The acceptance criteria are:

```text
Beautiful UI
+
Real API data
+
All major backend capabilities exposed
+
Working interactions
+
Working WebSocket updates
+
Correct empty/loading/error states
+
Consistent design
+
No fake data
+
No hardcoded metrics
+
No broken navigation
+
Production-quality UX
```

All of these are required.

---

# 29. FINAL AUDIT BEFORE YOU FINISH

Before saying the task is complete, inspect the application as if you are a real user.

Ask:

1. Does every sidebar item work?
2. Does every page use real API data?
3. Why would Memories ever show zero?
4. Can I search?
5. Can I filter?
6. Can I inspect an event?
7. Can I inspect a memory?
8. Can I inspect a session?
9. Can I interact with the graph?
10. Does graph search work?
11. Does graph filtering work?
12. Does WebSocket live updating work?
13. Are connection states accurate?
14. Do browser refreshes work?
15. Does browser back/forward work?
16. Are loading states polished?
17. Are errors handled?
18. Are empty states accurate?
19. Does the application look like one coherent product?
20. Does the UI match the quality and cleanliness of the uploaded SaaS reference?

If any answer is "no", continue fixing the application.

Do not stop at the first successful build.

---

## FINAL INSTRUCTION

**Take ownership of the entire frontend experience.**

Inspect first.

Understand the real APIs.

Fix the data flow.

Build a consistent design system.

Implement every missing UI.

Connect every existing API.

Fix broken pages.

Use real data.

Use the **last uploaded SaaS Startup screenshot as the primary visual inspiration**.

Then test the complete application end-to-end.

The final result should feel like a **finished professional Agent Ledger product**, not a collection of prototype screens.
