package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	AgentDirName = ".agent"
	ProjectMDFile = "project.md"
	StateDir = "state"
	CurrentStateFile = "current.json"
	DecisionsDir = "decisions"
	DiscoveriesDir = "discoveries"
	FailuresDir = "failures"
	ConstraintsDir = "constraints"
	SessionsDir = "sessions"
)

// Storage manages the .agent/ directory structure
type Storage struct {
	rootDir string
}

// New creates a new Storage instance for the given repository root
func New(repoRoot string) *Storage {
	return &Storage{
		rootDir: filepath.Join(repoRoot, AgentDirName),
	}
}

// Initialize creates the .agent/ directory structure
func (s *Storage) Initialize() error {
	// Create main .agent directory
	if err := os.MkdirAll(s.rootDir, 0755); err != nil {
		return fmt.Errorf("failed to create .agent directory: %w", err)
	}
	
	// Create subdirectories
	dirs := []string{
		filepath.Join(s.rootDir, StateDir),
		filepath.Join(s.rootDir, DecisionsDir),
		filepath.Join(s.rootDir, DiscoveriesDir),
		filepath.Join(s.rootDir, FailuresDir),
		filepath.Join(s.rootDir, ConstraintsDir),
		filepath.Join(s.rootDir, SessionsDir),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	return nil
}

// Exists checks if the .agent/ directory exists
func (s *Storage) Exists() bool {
	_, err := os.Stat(s.rootDir)
	return err == nil
}

// WriteJSON writes JSON data to a file
func (s *Storage) WriteJSON(path string, data interface{}) error {
	fullPath := filepath.Join(s.rootDir, path)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	if err := os.WriteFile(fullPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}
	
	return nil
}

// ReadJSON reads JSON data from a file
func (s *Storage) ReadJSON(path string, data interface{}) error {
	fullPath := filepath.Join(s.rootDir, path)
	
	jsonData, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}
	
	if err := json.Unmarshal(jsonData, data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	return nil
}

// WriteMarkdown writes Markdown content to a file
func (s *Storage) WriteMarkdown(path string, content string) error {
	fullPath := filepath.Join(s.rootDir, path)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}
	
	return nil
}

// ReadMarkdown reads Markdown content from a file
func (s *Storage) ReadMarkdown(path string) (string, error) {
	fullPath := filepath.Join(s.rootDir, path)
	
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}
	
	return string(content), nil
}

// FileExists checks if a file exists
func (s *Storage) FileExists(path string) bool {
	fullPath := filepath.Join(s.rootDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// ListFiles lists files in a directory
func (s *Storage) ListFiles(path string) ([]string, error) {
	fullPath := filepath.Join(s.rootDir, path)
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

// ListDirectories lists subdirectories in a directory
func (s *Storage) ListDirectories(path string) ([]string, error) {
	fullPath := filepath.Join(s.rootDir, path)
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}
	
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	
	return dirs, nil
}

// GetRoot returns the .agent/ directory path
func (s *Storage) GetRoot() string {
	return s.rootDir
}
