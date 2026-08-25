package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStorage(t *testing.T) {
	tempDir := t.TempDir()
	
	st := New(tempDir)
	if st == nil {
		t.Fatal("NewStorage returned nil")
	}
	
	expectedRoot := filepath.Join(tempDir, ".agent")
	if st.GetRoot() != expectedRoot {
		t.Errorf("Expected root %s, got %s", expectedRoot, st.GetRoot())
	}
}

func TestInitialize(t *testing.T) {
	tempDir := t.TempDir()
	st := New(tempDir)
	
	err := st.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Check that main directory was created
	if _, err := os.Stat(st.GetRoot()); os.IsNotExist(err) {
		t.Error(".agent directory was not created")
	}
	
	// Check that subdirectories were created
	expectedDirs := []string{
		filepath.Join(st.GetRoot(), "state"),
		filepath.Join(st.GetRoot(), "decisions"),
		filepath.Join(st.GetRoot(), "discoveries"),
		filepath.Join(st.GetRoot(), "failures"),
		filepath.Join(st.GetRoot(), "constraints"),
		filepath.Join(st.GetRoot(), "sessions"),
	}
	
	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", dir)
		}
	}
}

func TestExists(t *testing.T) {
	tempDir := t.TempDir()
	st := New(tempDir)
	
	// Should not exist initially
	if st.Exists() {
		t.Error("Storage should not exist before initialization")
	}
	
	// Initialize
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Should exist after initialization
	if !st.Exists() {
		t.Error("Storage should exist after initialization")
	}
}

func TestWriteReadJSON(t *testing.T) {
	tempDir := t.TempDir()
	st := New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	testData := struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}{
		Name:  "test",
		Value: 42,
	}
	
	// Write JSON
	err := st.WriteJSON("test/test.json", testData)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}
	
	// Read JSON
	var readData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	err = st.ReadJSON("test/test.json", &readData)
	if err != nil {
		t.Fatalf("ReadJSON failed: %v", err)
	}
	
	if readData.Name != testData.Name {
		t.Errorf("Expected name %s, got %s", testData.Name, readData.Name)
	}
	if readData.Value != testData.Value {
		t.Errorf("Expected value %d, got %d", testData.Value, readData.Value)
	}
}

func TestWriteReadMarkdown(t *testing.T) {
	tempDir := t.TempDir()
	st := New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	testContent := "# Test Document\n\nThis is a test."
	
	// Write Markdown
	err := st.WriteMarkdown("test/test.md", testContent)
	if err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}
	
	// Read Markdown
	readContent, err := st.ReadMarkdown("test/test.md")
	if err != nil {
		t.Fatalf("ReadMarkdown failed: %v", err)
	}
	
	if readContent != testContent {
		t.Errorf("Expected content %s, got %s", testContent, readContent)
	}
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()
	st := New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// File should not exist initially
	if st.FileExists("test/test.md") {
		t.Error("File should not exist before writing")
	}
	
	// Write file
	err := st.WriteMarkdown("test/test.md", "test content")
	if err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}
	
	// File should exist after writing
	if !st.FileExists("test/test.md") {
		t.Error("File should exist after writing")
	}
}

func TestListFiles(t *testing.T) {
	tempDir := t.TempDir()
	st := New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Create some test files
	err := st.WriteMarkdown("test/file1.md", "content1")
	if err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}
	err = st.WriteMarkdown("test/file2.md", "content2")
	if err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}
	
	// List files
	files, err := st.ListFiles("test")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestListDirectories(t *testing.T) {
	tempDir := t.TempDir()
	st := New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// List directories in sessions
	dirs, err := st.ListDirectories("sessions")
	if err != nil {
		t.Fatalf("ListDirectories failed: %v", err)
	}
	
	// Should be empty initially
	if len(dirs) != 0 {
		t.Errorf("Expected 0 directories, got %d", len(dirs))
	}
	
	// Create a session directory
	err = st.WriteJSON("sessions/test/metadata.json", map[string]string{"id": "test"})
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}
	
	// List directories again
	dirs, err = st.ListDirectories("sessions")
	if err != nil {
		t.Fatalf("ListDirectories failed: %v", err)
	}
	
	if len(dirs) != 1 {
		t.Errorf("Expected 1 directory, got %d", len(dirs))
	}
}
