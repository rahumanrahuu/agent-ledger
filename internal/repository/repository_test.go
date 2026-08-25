package repository

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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

	// Create an initial commit so HEAD exists
	testFile := filepath.Join(tempDir, "init.txt")
	if err := os.WriteFile(testFile, []byte("init content"), 0644); err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	cmd = exec.Command("git", "add", "init.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add initial file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit initial file: %v", err)
	}

	return tempDir
}

func TestMustBeInRepository(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Inside git repository
	repoDir := setupTestRepo(t)
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("Failed to chdir to repo: %v", err)
	}

	if err := MustBeInRepository(); err != nil {
		t.Errorf("Expected nil error in git repository, got: %v", err)
	}

	// Outside git repository
	nonRepoDir := t.TempDir()
	if err := os.Chdir(nonRepoDir); err != nil {
		t.Fatalf("Failed to chdir to non-repo: %v", err)
	}

	if err := MustBeInRepository(); err == nil {
		t.Errorf("Expected error when outside git repository, got nil")
	}
}

func TestGetRepositoryRoot(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	repoDir := setupTestRepo(t)
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("Failed to chdir to repo: %v", err)
	}

	root, err := GetRepositoryRoot()
	if err != nil {
		t.Fatalf("GetRepositoryRoot failed: %v", err)
	}

	if !strings.HasSuffix(root, repoDir) && !strings.HasSuffix(repoDir, root) {
		t.Errorf("Expected root %s to match %s", root, repoDir)
	}
}

func TestDetect(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// In non-repo directory
	nonRepoDir := t.TempDir()
	if err := os.Chdir(nonRepoDir); err != nil {
		t.Fatalf("Failed to chdir to non-repo: %v", err)
	}
	if _, err := Detect(); err == nil {
		t.Error("Detect should fail outside git repository")
	}

	// In clean repo
	repoDir := setupTestRepo(t)
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("Failed to chdir to repo: %v", err)
	}

	repo, err := Detect()
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if repo.Dirty {
		t.Error("Expected repository to be clean")
	}
	if repo.Branch != "main" && repo.Branch != "master" {
		t.Errorf("Unexpected branch name: %s", repo.Branch)
	}
	if repo.Head == "" {
		t.Error("Expected non-empty HEAD commit")
	}
	if len(repo.Staged) != 0 || len(repo.Unstaged) != 0 || len(repo.Untracked) != 0 {
		t.Errorf("Expected no status changes in clean repo, got staged=%v unstaged=%v untracked=%v",
			repo.Staged, repo.Unstaged, repo.Untracked)
	}

	// Add an untracked file
	untrackedFile := filepath.Join(repoDir, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("untracked"), 0644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	repo, err = Detect()
	if err != nil {
		t.Fatalf("Detect failed with untracked file: %v", err)
	}
	if !repo.Dirty {
		t.Error("Expected repo to be dirty after creating untracked file")
	}
	if len(repo.Untracked) == 0 || repo.Untracked[0] != "untracked.txt" {
		t.Errorf("Expected untracked.txt in Untracked, got %v", repo.Untracked)
	}

	// Stage a new file
	stagedFile := filepath.Join(repoDir, "staged.txt")
	if err := os.WriteFile(stagedFile, []byte("staged"), 0644); err != nil {
		t.Fatalf("Failed to create staged file: %v", err)
	}
	cmd := exec.Command("git", "add", "staged.txt")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add staged file: %v", err)
	}

	repo, err = Detect()
	if err != nil {
		t.Fatalf("Detect failed with staged file: %v", err)
	}
	if !repo.Dirty {
		t.Error("Expected repo to be dirty after staging file")
	}
	foundStaged := false
	for _, f := range repo.Staged {
		if f == "staged.txt" {
			foundStaged = true
			break
		}
	}
	if !foundStaged {
		t.Errorf("Expected staged.txt in Staged list, got %v", repo.Staged)
	}

	// Modify tracked file (init.txt) without staging
	initFile := filepath.Join(repoDir, "init.txt")
	if err := os.WriteFile(initFile, []byte("modified content"), 0644); err != nil {
		t.Fatalf("Failed to modify init.txt: %v", err)
	}

	repo, err = Detect()
	if err != nil {
		t.Fatalf("Detect failed with modified file: %v", err)
	}
	foundUnstaged := false
	for _, f := range repo.Unstaged {
		if f == "init.txt" {
			foundUnstaged = true
			break
		}
	}
	if !foundUnstaged {
		t.Errorf("Expected init.txt in Unstaged list, got %v", repo.Unstaged)
	}
}
