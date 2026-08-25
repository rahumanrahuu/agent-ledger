package checkpoint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	
	"agent-ledger/internal/git"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
)

func setupTestGitRepo(t *testing.T) string {
	tempDir := t.TempDir()
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}
	
	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user: %v", err)
	}
	
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git email: %v", err)
	}
	
	// Create initial commit
	testFile := filepath.Join(tempDir, "initial.txt")
	if err := os.WriteFile(testFile, []byte("initial content"), 0644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}
	
	cmd = exec.Command("git", "add", "initial.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add initial file: %v", err)
	}
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}
	
	return tempDir
}

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()
	st := storage.New(tempDir)
	
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	manager := NewManager(st)
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestCreateCheckpoint(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestGitRepo(t)
	os.Chdir(gitDir)
	
	// Initialize storage
	st := storage.New(gitDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Create a session
	sessionID := "test-session-123"
	sessionDir := filepath.Join(gitDir, ".agent", "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	
	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		t.Fatalf("Failed to detect repository: %v", err)
	}
	
	// Override repo root to match test directory
	repo.Root = gitDir
	
	// Create checkpoint
	manager := NewManager(st)
	checkpoint, err := manager.Create(sessionID, repo)
	if err != nil {
		t.Fatalf("Create checkpoint failed: %v", err)
	}
	
	if checkpoint.ID != "checkpoint-1" {
		t.Errorf("Expected checkpoint ID checkpoint-1, got %s", checkpoint.ID)
	}
	if checkpoint.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, checkpoint.SessionID)
	}
	if checkpoint.Commit == "" {
		t.Error("Checkpoint commit should not be empty")
	}
	if checkpoint.Tree == "" {
		t.Error("Checkpoint tree should not be empty")
	}
	
	// Verify the checkpoint ref exists
	refPath := filepath.Join(gitDir, ".git", "refs", "agents", "sessions", sessionID, "checkpoint-1")
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		t.Errorf("Checkpoint ref should exist at %s", refPath)
	}
}

func TestCreateCheckpointWithStagedAndUnstagedChanges(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestGitRepo(t)
	os.Chdir(gitDir)
	
	// Initialize storage
	st := storage.New(gitDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Create a session
	sessionID := "test-session-staged"
	sessionDir := filepath.Join(gitDir, ".agent", "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	
	// Create staged changes
	testFile := filepath.Join(gitDir, "staged.txt")
	if err := os.WriteFile(testFile, []byte("staged content"), 0644); err != nil {
		t.Fatalf("Failed to create staged file: %v", err)
	}
	
	cmd := exec.Command("git", "add", "staged.txt")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to stage file: %v", err)
	}
	
	// Create unstaged changes
	if err := os.WriteFile(testFile, []byte("staged content\nunstaged change"), 0644); err != nil {
		t.Fatalf("Failed to modify staged file: %v", err)
	}
	
	// Create untracked file
	untrackedFile := filepath.Join(gitDir, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("untracked content"), 0644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}
	
	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		t.Fatalf("Failed to detect repository: %v", err)
	}
	
	// Override repo root to match test directory
	repo.Root = gitDir
	
	// Verify git state before checkpoint
	beforeStaged := repo.Staged
	beforeUnstaged := repo.Unstaged
	beforeUntracked := repo.Untracked
	
	// Create checkpoint
	manager := NewManager(st)
	checkpoint, err := manager.Create(sessionID, repo)
	if err != nil {
		t.Fatalf("Create checkpoint failed: %v", err)
	}
	
	// Verify git state after checkpoint is preserved (except for .agent/ directory creation)
	statusAfter, err := git.GetStatus()
	if err != nil {
		t.Fatalf("Failed to get status after checkpoint: %v", err)
	}
	
	// The status should be mostly the same (index-preserving), but .agent/ directory may be added
	// The important thing is that staged/unstaged state of the original files is preserved
	if !strings.Contains(statusAfter, "AM staged.txt") {
		t.Errorf("Staged file state should be preserved. Status after: %s", statusAfter)
	}
	if !strings.Contains(statusAfter, "?? untracked.txt") {
		t.Errorf("Untracked file should still be present. Status after: %s", statusAfter)
	}
	
	// Verify checkpoint captured the changes
	if len(checkpoint.ChangedFiles) == 0 {
		t.Error("Checkpoint should capture changed files")
	}
	
	// Verify checkpoint metadata
	if len(checkpoint.GitStatus.StagedFiles) != len(beforeStaged) {
		t.Errorf("Checkpoint should capture staged files. Expected %d, got %d", len(beforeStaged), len(checkpoint.GitStatus.StagedFiles))
	}
	
	if len(checkpoint.GitStatus.UntrackedFiles) != len(beforeUntracked) {
		t.Errorf("Checkpoint should capture untracked files. Expected %d, got %d", len(beforeUntracked), len(checkpoint.GitStatus.UntrackedFiles))
	}
	
	t.Logf("✓ Index-preserving checkpoint test passed")
	t.Logf("  - Staged files: %v", beforeStaged)
	t.Logf("  - Unstaged files: %v", beforeUnstaged)
	t.Logf("  - Untracked files: %v", beforeUntracked)
	t.Logf("  - Checkpoint commit: %s", checkpoint.Commit)
}

func TestListCheckpoints(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestGitRepo(t)
	os.Chdir(gitDir)
	
	// Initialize storage
	st := storage.New(gitDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Create a session
	sessionID := "test-session-list"
	sessionDir := filepath.Join(gitDir, ".agent", "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	
	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		t.Fatalf("Failed to detect repository: %v", err)
	}
	
	// Override repo root to match test directory
	repo.Root = gitDir
	
	// Create multiple checkpoints
	manager := NewManager(st)
	
	_, err = manager.Create(sessionID, repo)
	if err != nil {
		t.Fatalf("Create first checkpoint failed: %v", err)
	}
	
	// Make a change
	testFile := filepath.Join(gitDir, "change1.txt")
	if err := os.WriteFile(testFile, []byte("change 1"), 0644); err != nil {
		t.Fatalf("Failed to create change file: %v", err)
	}
	
	_, err = manager.Create(sessionID, repo)
	if err != nil {
		t.Fatalf("Create second checkpoint failed: %v", err)
	}
	
	// List checkpoints
	checkpoints, err := manager.List(sessionID)
	if err != nil {
		t.Fatalf("List checkpoints failed: %v", err)
	}
	
	if len(checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints, got %d", len(checkpoints))
	}
	
	if checkpoints[0].ID != "checkpoint-1" {
		t.Errorf("Expected first checkpoint ID checkpoint-1, got %s", checkpoints[0].ID)
	}
	
	if checkpoints[1].ID != "checkpoint-2" {
		t.Errorf("Expected second checkpoint ID checkpoint-2, got %s", checkpoints[1].ID)
	}
}

func TestGetCheckpoint(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestGitRepo(t)
	os.Chdir(gitDir)
	
	// Initialize storage
	st := storage.New(gitDir)
	if err := st.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Create a session
	sessionID := "test-session-get"
	sessionDir := filepath.Join(gitDir, ".agent", "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	
	// Get repository info
	repo, err := repository.Detect()
	if err != nil {
		t.Fatalf("Failed to detect repository: %v", err)
	}
	
	// Override repo root to match test directory
	repo.Root = gitDir
	
	// Create checkpoint
	manager := NewManager(st)
	createdCheckpoint, err := manager.Create(sessionID, repo)
	if err != nil {
		t.Fatalf("Create checkpoint failed: %v", err)
	}
	
	// Get checkpoint
	retrievedCheckpoint, err := manager.Get(sessionID, createdCheckpoint.ID)
	if err != nil {
		t.Fatalf("Get checkpoint failed: %v", err)
	}
	
	if retrievedCheckpoint.ID != createdCheckpoint.ID {
		t.Errorf("Expected checkpoint ID %s, got %s", createdCheckpoint.ID, retrievedCheckpoint.ID)
	}
	
	if retrievedCheckpoint.Commit != createdCheckpoint.Commit {
		t.Errorf("Expected commit %s, got %s", createdCheckpoint.Commit, retrievedCheckpoint.Commit)
	}
}
