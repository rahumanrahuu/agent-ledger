package api

import (
	"testing"
	"time"

	"agent-ledger/internal/memory"
)

func newMemoryAPI(t *testing.T) (*API, *memory.Manager) {
	t.Helper()
	mgr, err := memory.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create memory manager: %v", err)
	}
	t.Cleanup(func() { _ = mgr.Close() })
	return &API{memoryMgr: mgr}, mgr
}

func addMemory(t *testing.T, mgr *memory.Manager, mem memory.Memory) {
	t.Helper()
	if mem.CreatedAt.IsZero() {
		mem.CreatedAt = time.Now()
	}
	if err := mgr.Add(mem); err != nil {
		t.Fatalf("add memory %q: %v", mem.ID, err)
	}
}

func TestGetMemoriesListsAndFilters(t *testing.T) {
	api, mgr := newMemoryAPI(t)
	addMemory(t, mgr, memory.Memory{ID: "decision-1", Type: "decision", Title: "Use Go"})
	addMemory(t, mgr, memory.Memory{ID: "discovery-1", Type: "discovery", Title: "Fast builds"})

	all, err := api.GetMemories("", "all", 20)
	if err != nil || len(all) != 2 {
		t.Fatalf("list all: got %d memories, err=%v", len(all), err)
	}
	decisions, err := api.GetMemories("", "decision", 20)
	if err != nil || len(decisions) != 1 || decisions[0].ID != "decision-1" {
		t.Fatalf("filter decisions: got %#v, err=%v", decisions, err)
	}
}

func TestGetMemoriesSearchUsesStableShape(t *testing.T) {
	api, mgr := newMemoryAPI(t)
	addMemory(t, mgr, memory.Memory{ID: "search-1", Type: "decision", Title: "Backend language", Content: "Use Go for the backend", Keywords: "go,backend"})

	results, err := api.GetMemories("backend", "all", 10)
	if err != nil {
		t.Fatalf("search memories: %v", err)
	}
	if len(results) != 1 || results[0].ID != "search-1" {
		t.Fatalf("unexpected search results: %#v", results)
	}
}

func TestGetMemoriesHandlesDisabledManager(t *testing.T) {
	api := &API{}
	results, err := api.GetMemories("", "all", 20)
	if err != nil || results == nil || len(results) != 0 {
		t.Fatalf("disabled manager: got %#v, err=%v", results, err)
	}
}

func TestGetMemoriesAppliesDefaultLimit(t *testing.T) {
	api, mgr := newMemoryAPI(t)
	for i := 0; i < 25; i++ {
		addMemory(t, mgr, memory.Memory{ID: string(rune('a' + i)), Type: "decision", Title: "Memory"})
	}
	results, err := api.GetMemories("", "all", 0)
	if err != nil || len(results) != 20 {
		t.Fatalf("default limit: got %d memories, err=%v", len(results), err)
	}
}
