package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"strings"
)

func setupTestRepo(t *testing.T) string {
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
	
	return tempDir
}

func TestIsRepository(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	// Test in a git repository
	gitDir := setupTestRepo(t)
	os.Chdir(gitDir)
	
	isRepo, err := IsRepository()
	if err != nil {
		t.Fatalf("IsRepository failed: %v", err)
	}
	if !isRepo {
		t.Error("Should be a git repository")
	}
	
	// Test outside git repository
	nonGitDir := t.TempDir()
	os.Chdir(nonGitDir)
	
	isRepo, err = IsRepository()
	if err != nil {
		t.Fatalf("IsRepository failed: %v", err)
	}
	if isRepo {
		t.Error("Should not be a git repository")
	}
}

func TestGetRepositoryRoot(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestRepo(t)
	os.Chdir(gitDir)
	
	root, err := GetRepositoryRoot()
	if err != nil {
		t.Fatalf("GetRepositoryRoot failed: %v", err)
	}
	
	// The root should be the git directory we created (normalize path comparison)
	if !strings.HasSuffix(root, gitDir) && !strings.HasSuffix(gitDir, root) {
		t.Errorf("Expected root to match %s, got %s", gitDir, root)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestRepo(t)
	os.Chdir(gitDir)
	
	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	
	// Default branch should be main or master
	if branch != "main" && branch != "master" {
		t.Errorf("Expected branch to be main or master, got %s", branch)
	}
}

func TestGetHeadCommit(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestRepo(t)
	os.Chdir(gitDir)
	
	// Create a commit
	testFile := filepath.Join(gitDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	
	cmd = exec.Command("git", "commit", "-m", "Test commit")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}
	
	// Get HEAD commit
	head, err := GetHeadCommit()
	if err != nil {
		t.Fatalf("GetHeadCommit failed: %v", err)
	}
	
	if head == "" {
		t.Error("HEAD commit should not be empty")
	}
}

func TestIsDirty(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestRepo(t)
	os.Chdir(gitDir)
	
	// Initial state should be clean
	dirty, err := IsDirty()
	if err != nil {
		t.Fatalf("IsDirty failed: %v", err)
	}
	if dirty {
		t.Error("Repository should be clean initially")
	}
	
	// Create a file without committing
	testFile := filepath.Join(gitDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Now should be dirty
	dirty, err = IsDirty()
	if err != nil {
		t.Fatalf("IsDirty failed: %v", err)
	}
	if !dirty {
		t.Error("Repository should be dirty after creating untracked file")
	}
}

func TestWriteTree(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestRepo(t)
	os.Chdir(gitDir)
	
	// Create and stage a file
	testFile := filepath.Join(gitDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	
	// Write tree
	tree, err := WriteTree()
	if err != nil {
		t.Fatalf("WriteTree failed: %v", err)
	}
	
	if tree == "" {
		t.Error("Tree should not be empty")
	}
}

func TestCommitTree(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestRepo(t)
	os.Chdir(gitDir)
	
	// Create and stage a file
	testFile := filepath.Join(gitDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	
	// Get tree
	tree, err := WriteTree()
	if err != nil {
		t.Fatalf("WriteTree failed: %v", err)
	}
	
	// Create commit
	commit, err := CommitTree("Test commit", "", tree)
	if err != nil {
		t.Fatalf("CommitTree failed: %v", err)
	}
	
	if commit == "" {
		t.Error("Commit should not be empty")
	}
}

func TestUpdateRef(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	gitDir := setupTestRepo(t)
	os.Chdir(gitDir)
	
	// Create a commit first
	testFile := filepath.Join(gitDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	
	cmd = exec.Command("git", "commit", "-m", "Test commit")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}
	
	head, err := GetHeadCommit()
	if err != nil {
		t.Fatalf("GetHeadCommit failed: %v", err)
	}
	
	// Create a test ref
	err = UpdateRef("refs/test/test-ref", head)
	if err != nil {
		t.Fatalf("UpdateRef failed: %v", err)
	}
	
	// Verify ref exists
	refContent, err := ShowRef("refs/test/test-ref")
	if err != nil {
		t.Fatalf("ShowRef failed: %v", err)
	}
	
	if refContent == "" {
		t.Error("Ref should exist after UpdateRef")
	}
}
