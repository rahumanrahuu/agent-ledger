package ui

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"agent-ledger/internal/api"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
	"github.com/gorilla/websocket"
)

// Server handles the UI HTTP server
type Server struct {
	repo    *repository.Repository
	storage *storage.Storage
	port    int
	host    string
	version string
	api     *api.API
	clients map[*websocket.Conn]struct{}
	mu      sync.Mutex
}

// NewServer creates a new UI server
func NewServer(repo *repository.Repository, storage *storage.Storage, port int, host string, version string) *Server {
	return &Server{
		repo:    repo,
		storage: storage,
		port:    port,
		host:    host,
		version: version,
		api:     api.NewAPI(repo, storage, version),
		clients: make(map[*websocket.Conn]struct{}),
	}
}

// Start starts the UI server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/overview", s.handleOverview)
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/sessions/", s.handleSessionDetail)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/graph", s.handleGraph)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/memories", s.handleMemories)
	mux.HandleFunc("/api/live", s.handleLive)

	// Frontend - will be served from embedded assets in the future
	// For now, serve index.html for SPA routing
	mux.HandleFunc("/", s.handleFrontend)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w. Port may be in use. Try a different port with --port flag.", addr, err)
	}
	defer listener.Close()

	url := fmt.Sprintf("http://%s", addr)
	fmt.Printf("Agent Ledger UI v%s starting at %s\n", s.version, url)

	// Check if embedded assets are available
	_, embedErr := fs.Sub(Dist, "dist")
	if embedErr == nil {
		fmt.Println("Embedded UI assets: available")
	} else {
		fmt.Println("Embedded UI assets: not available (development mode)")
	}

	fmt.Println("Press Ctrl+C to stop")
	go s.watchLedger()

	return http.Serve(listener, mux)
}

// handleMemories serves a stable list response for both browsing and searching
// the memory store.
func (s *Server) handleMemories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	memories, err := s.api.GetMemories(r.URL.Query().Get("q"), r.URL.Query().Get("type"), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memories)
}

var liveUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// handleLive keeps browsers connected for lightweight ledger-change notifications.
// API responses remain the source of truth; clients refetch after each notification.
func (s *Server) handleLive(w http.ResponseWriter, r *http.Request) {
	conn, err := liveUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.clients[conn] = struct{}{}
	s.mu.Unlock()
	_ = conn.WriteJSON(map[string]any{"type": "connected", "timestamp": time.Now().UTC()})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			_ = conn.Close()
			return
		}
	}
}

// watchLedger polls a small metadata fingerprint. This catches writes from the
// CLI and the separate MCP process without coupling either writer to the UI.
func (s *Server) watchLedger() {
	last := s.ledgerFingerprint()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		next := s.ledgerFingerprint()
		if next != last {
			last = next
			s.broadcastLiveUpdate()
		}
	}
}

func (s *Server) ledgerFingerprint() string {
	root := filepath.Join(s.repo.Root, ".agent")
	entries := make([]string, 0)
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := os.Stat(path)
		if err == nil {
			entries = append(entries, fmt.Sprintf("%s:%d:%d", path, info.Size(), info.ModTime().UnixNano()))
		}
		return nil
	})
	sort.Strings(entries)
	return strings.Join(entries, "|")
}

func (s *Server) broadcastLiveUpdate() {
	payload, _ := json.Marshal(map[string]any{"type": "ledger.updated", "timestamp": time.Now().UTC()})
	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			_ = conn.Close()
			delete(s.clients, conn)
		}
	}
}

// handleOverview serves the project overview
func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	overview, err := s.api.GetOverview()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(overview)
}

// handleSessions serves the sessions list
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if this is a specific session detail request
	if strings.HasPrefix(r.URL.Path, "/api/sessions/") {
		sessionID := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
		if sessionID != "" && sessionID != "/" {
			detail, err := s.api.GetSessionDetail(sessionID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(detail)
			return
		}
	}

	sessions, err := s.api.GetSessions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// handleSessionDetail serves a specific session's details
func (s *Server) handleSessionDetail(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/sessions/")

	if sessionID == "" || sessionID == "/" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	detail, err := s.api.GetSessionDetail(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// handleEvents serves the events timeline
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filterType := r.URL.Query().Get("type")
	events, err := s.api.GetEvents(filterType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// handleGraph serves the knowledge graph
func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	graph, err := s.api.GetGraph()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graph)
}

// handleSearch serves search results
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query required", http.StatusBadRequest)
		return
	}

	results, err := s.api.Search(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// handleFrontend serves the frontend application
func (s *Server) handleFrontend(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Block serving the API routes through this handler
	if strings.HasPrefix(path, "/api/") {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Try to serve from embedded filesystem
	fsys, err := fs.Sub(Dist, "dist")
	if err != nil {
		// Embedded assets not available - this is a build error
		http.Error(w, "UI assets not embedded. Rebuild with frontend: cd ui/frontend && npm install && npm run build", http.StatusInternalServerError)
		return
	}

	// Embedded assets are available
	if path == "/" {
		path = "index.html"
	} else {
		path = strings.TrimPrefix(path, "/")
	}

	file, err := fs.ReadFile(fsys, path)
	if err != nil {
		// Try index.html for SPA routing
		file, err = fs.ReadFile(fsys, "index.html")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}

	// Set content type based on file extension
	contentType := "text/html; charset=utf-8"
	if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript"
	} else if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(path, ".json") {
		contentType = "application/json"
	} else if strings.HasSuffix(path, ".svg") {
		contentType = "image/svg+xml"
	} else if strings.HasSuffix(path, ".png") {
		contentType = "image/png"
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(file)
}
