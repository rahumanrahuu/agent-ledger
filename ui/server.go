package ui

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"

	"agent-ledger/internal/api"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
)

// UI is the embedded frontend filesystem (empty in dev builds)
var UI fs.FS = nil

// Server handles the UI HTTP server
type Server struct {
	repo    *repository.Repository
	storage *storage.Storage
	port    int
	version string
	api     *api.API
}

// NewServer creates a new UI server
func NewServer(repo *repository.Repository, storage *storage.Storage, port int, version string) *Server {
	return &Server{
		repo:    repo,
		storage: storage,
		port:    port,
		version: version,
		api:     api.NewAPI(repo, storage, version),
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

	// Frontend - will be served from embedded assets in the future
	// For now, serve index.html for SPA routing
	mux.HandleFunc("/", s.handleFrontend)

	addr := fmt.Sprintf("localhost:%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	defer listener.Close()

	url := fmt.Sprintf("http://%s", addr)
	fmt.Printf("Agent Ledger UI starting at %s\n", url)
	fmt.Println("Press Ctrl+C to stop")

	return http.Serve(listener, mux)
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

	// If UI is embedded, serve from it
	if UI != nil {
		// Serve files from embedded filesystem
		if path == "/" {
			path = "index.html"
		} else {
			path = strings.TrimPrefix(path, "/")
		}

		file, err := fs.ReadFile(UI, path)
		if err != nil {
			// Try index.html for SPA routing
			file, err = fs.ReadFile(UI, "index.html")
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
		return
	}

	// Development mode: serve development message
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Agent Ledger UI</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
            background: #f5f5f7;
            color: #333;
            padding: 40px;
            line-height: 1.6;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            padding: 40px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        h1 {
            font-size: 32px;
            font-weight: 600;
            margin-bottom: 10px;
        }
        p {
            color: #666;
            margin-bottom: 20px;
        }
        .info-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 20px;
            margin-top: 30px;
        }
        .info-card {
            border: 1px solid #e5e5e7;
            border-radius: 8px;
            padding: 20px;
        }
        .info-card h3 {
            font-size: 14px;
            font-weight: 500;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 8px;
        }
        .info-card p {
            font-size: 16px;
            margin: 0;
            font-family: "SF Mono", Monaco, "Cascadia Code", monospace;
            word-break: break-all;
        }
        .status {
            background: #f5f5f7;
            border-left: 4px solid #34c759;
            padding: 15px;
            border-radius: 4px;
            margin-top: 20px;
            font-size: 14px;
            color: #333;
        }
        .status strong {
            color: #34c759;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Agent Ledger UI</h1>
        <p>Local web interface for exploring agent sessions, decisions, discoveries, and checkpoints.</p>

        <div class="status">
            <strong>✓ Server running</strong> – API endpoints are available at /api/*
        </div>

        <div class="info-grid">
            <div class="info-card">
                <h3>API Base</h3>
                <p>/api</p>
            </div>
            <div class="info-card">
                <h3>Overview</h3>
                <p>/api/overview</p>
            </div>
            <div class="info-card">
                <h3>Sessions</h3>
                <p>/api/sessions</p>
            </div>
            <div class="info-card">
                <h3>Events</h3>
                <p>/api/events</p>
            </div>
            <div class="info-card">
                <h3>Graph</h3>
                <p>/api/graph</p>
            </div>
            <div class="info-card">
                <h3>Search</h3>
                <p>/api/search?q=query</p>
            </div>
        </div>

        <p style="margin-top: 30px; font-size: 12px; color: #999;">
            Frontend assets will be embedded in the binary upon build completion.
        </p>
    </div>
</body>
</html>`)
}
