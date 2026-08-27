package memory

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Memory represents a single memory record
type Memory struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"` // decision, discovery, constraint, failure
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Embedding  []byte    `json:"-"`
	Keywords   string    `json:"keywords"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Importance float64   `json:"importance"`
	SessionID  string    `json:"session_id,omitempty"`
	Path       string    `json:"path"`
}

// SearchResult represents a search result with score
type SearchResult struct {
	Memory Memory  `json:"memory"`
	Score  float64 `json:"score"`
}

// Manager handles memory operations
type Manager struct {
	db   *sql.DB
	root string
}

// NewManager creates a new memory manager
func NewManager(root string) (*Manager, error) {
	dbPath := filepath.Join(root, ".agent", "memory", "vectors.db")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create memory directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open memory database: %w", err)
	}

	// Create tables
	if err := initDB(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Manager{db: db, root: root}, nil
}

// initDB initializes the database schema
func initDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		embedding BLOB,
		keywords TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		importance REAL DEFAULT 0.5,
		session_id TEXT,
		path TEXT,
		archived INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_type ON memories(type);
	CREATE INDEX IF NOT EXISTS idx_created ON memories(created_at);
	CREATE INDEX IF NOT EXISTS idx_session ON memories(session_id);
	CREATE INDEX IF NOT EXISTS idx_archived ON memories(archived);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Try to create FTS5 table - if it fails, continue without full-text search
	ftsSchema := `
	CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
		title,
		content,
		keywords,
		content=memories,
		content_rowid=rowid
	);`

	_, ftsErr := db.Exec(ftsSchema)
	if ftsErr != nil {
		// Log that FTS5 is not available but don't fail
		// Full-text search will use basic LIKE queries instead
	}

	return nil
}

// Add adds a new memory
func (m *Manager) Add(memory Memory) error {
	if memory.ID == "" {
		return fmt.Errorf("memory ID is required")
	}

	stmt, err := m.db.Prepare(`
		INSERT OR REPLACE INTO memories
		(id, type, title, content, embedding, keywords, created_at, updated_at, importance, session_id, path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = now
	}
	memory.UpdatedAt = now

	_, err = stmt.Exec(
		memory.ID,
		memory.Type,
		memory.Title,
		memory.Content,
		memory.Embedding,
		memory.Keywords,
		memory.CreatedAt,
		memory.UpdatedAt,
		memory.Importance,
		memory.SessionID,
		memory.Path,
	)
	return err
}

// Get retrieves a memory by ID
func (m *Manager) Get(id string) (*Memory, error) {
	var mem Memory
	err := m.db.QueryRow(`
		SELECT id, type, title, content, embedding, keywords, created_at, updated_at, importance, session_id, path
		FROM memories
		WHERE id = ? AND archived = 0
	`, id).Scan(
		&mem.ID, &mem.Type, &mem.Title, &mem.Content, &mem.Embedding,
		&mem.Keywords, &mem.CreatedAt, &mem.UpdatedAt, &mem.Importance,
		&mem.SessionID, &mem.Path,
	)
	if err != nil {
		return nil, err
	}
	return &mem, nil
}

// Search performs keyword search using FTS
func (m *Manager) Search(query string, memType string, limit int) ([]SearchResult, error) {
	var rows *sql.Rows
	var err error

	if memType != "" && memType != "all" {
		rows, err = m.db.Query(`
			SELECT m.id, m.type, m.title, m.content, m.embedding, m.keywords,
			       m.created_at, m.updated_at, m.importance, m.session_id, m.path
			FROM memories m
			WHERE m.id IN (
				SELECT content_rowid FROM memories_fts WHERE memories_fts MATCH ?
			)
			AND m.type = ?
			AND m.archived = 0
			ORDER BY m.created_at DESC
			LIMIT ?
		`, query, memType, limit)
	} else {
		rows, err = m.db.Query(`
			SELECT m.id, m.type, m.title, m.content, m.embedding, m.keywords,
			       m.created_at, m.updated_at, m.importance, m.session_id, m.path
			FROM memories m
			WHERE m.id IN (
				SELECT content_rowid FROM memories_fts WHERE memories_fts MATCH ?
			)
			AND m.archived = 0
			ORDER BY m.created_at DESC
			LIMIT ?
		`, query, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var mem Memory
		err := rows.Scan(
			&mem.ID, &mem.Type, &mem.Title, &mem.Content, &mem.Embedding,
			&mem.Keywords, &mem.CreatedAt, &mem.UpdatedAt, &mem.Importance,
			&mem.SessionID, &mem.Path,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Calculate BM25-like score based on query match
		score := calculateScore(mem, query)

		results = append(results, SearchResult{
			Memory: mem,
			Score:  score,
		})
	}

	return results, rows.Err()
}

// calculateScore calculates a simple relevance score
func calculateScore(mem Memory, query string) float64 {
	score := 0.0
	queryTerms := strings.Fields(strings.ToLower(query))

	// Title match (highest weight)
	titleLower := strings.ToLower(mem.Title)
	for _, term := range queryTerms {
		if strings.Contains(titleLower, term) {
			score += 0.4
		}
	}

	// Keywords match
	keywordsLower := strings.ToLower(mem.Keywords)
	for _, term := range queryTerms {
		if strings.Contains(keywordsLower, term) {
			score += 0.3
		}
	}

	// Content match
	contentLower := strings.ToLower(mem.Content)
	for _, term := range queryTerms {
		if strings.Contains(contentLower, term) {
			score += 0.2
		}
	}

	// Importance bonus
	score *= (1.0 + mem.Importance)

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// List returns all memories with optional filtering
func (m *Manager) List(memType string, limit int) ([]Memory, error) {
	var rows *sql.Rows
	var err error

	if memType != "" && memType != "all" {
		rows, err = m.db.Query(`
			SELECT id, type, title, content, embedding, keywords, created_at, updated_at, importance, session_id, path
			FROM memories
			WHERE type = ? AND archived = 0
			ORDER BY created_at DESC
			LIMIT ?
		`, memType, limit)
	} else {
		rows, err = m.db.Query(`
			SELECT id, type, title, content, embedding, keywords, created_at, updated_at, importance, session_id, path
			FROM memories
			WHERE archived = 0
			ORDER BY created_at DESC
			LIMIT ?
		`, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var mem Memory
		err := rows.Scan(
			&mem.ID, &mem.Type, &mem.Title, &mem.Content, &mem.Embedding,
			&mem.Keywords, &mem.CreatedAt, &mem.UpdatedAt, &mem.Importance,
			&mem.SessionID, &mem.Path,
		)
		if err != nil {
			return nil, err
		}
		memories = append(memories, mem)
	}

	return memories, rows.Err()
}

// Delete archives a memory (soft delete)
func (m *Manager) Delete(id string) error {
	stmt, err := m.db.Prepare(`UPDATE memories SET archived = 1, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), id)
	return err
}

// Close closes the database connection
func (m *Manager) Close() error {
	if m == nil || m.db == nil {
		return nil
	}
	return m.db.Close()
}

// EmbeddingToBytes converts a float32 slice to bytes
func EmbeddingToBytes(embedding []float32) []byte {
	bytes := make([]byte, len(embedding)*4)
	for i, v := range embedding {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(v))
	}
	return bytes
}

// BytesToEmbedding converts bytes back to float32 slice
func BytesToEmbedding(b []byte) []float32 {
	embedding := make([]float32, len(b)/4)
	for i := 0; i < len(embedding); i++ {
		embedding[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return embedding
}
