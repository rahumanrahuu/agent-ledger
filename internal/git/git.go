package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Command executes a git command and returns the output
func Command(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w (stderr: %s)", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.String(), nil
}

// CommandInDir executes a git command in a specific directory
func CommandInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s failed in %s: %w (stderr: %s)", strings.Join(args, " "), dir, err, stderr.String())
	}

	return stdout.String(), nil
}

// IsRepository checks if the current directory is inside a git repository
func IsRepository() (bool, error) {
	_, err := Command("rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false, nil
	}
	return true, nil
}

// IsRepositoryInDir checks if the given directory is inside a git repository
func IsRepositoryInDir(dir string) (bool, error) {
	_, err := CommandInDir(dir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetRepositoryRoot returns the root directory of the git repository
func GetRepositoryRoot() (string, error) {
	output, err := Command("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return filepath.Clean(strings.TrimSpace(output)), nil
}

// GetRepositoryRootInDir returns the root directory of the git repository containing dir
func GetRepositoryRootInDir(dir string) (string, error) {
	output, err := CommandInDir(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return filepath.Clean(strings.TrimSpace(output)), nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch() (string, error) {
	output, err := Command("branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetHeadCommit returns the current HEAD commit SHA
func GetHeadCommit() (string, error) {
	output, err := Command("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetRemotes returns the configured remotes
func GetRemotes() (map[string]string, error) {
	output, err := Command("remote", "-v")
	if err != nil {
		return nil, err
	}

	remotes := make(map[string]string)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			url := parts[1]
			if _, exists := remotes[name]; !exists {
				remotes[name] = url
			}
		}
	}

	return remotes, nil
}

// IsDirty checks if there are uncommitted changes
func IsDirty() (bool, error) {
	output, err := Command("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

// GetStatus returns detailed git status information
func GetStatus() (string, error) {
	return Command("status", "--porcelain")
}

// GetDiff returns the diff of the working tree
func GetDiff() (string, error) {
	return Command("diff")
}

// GetStagedDiff returns the diff of staged changes
func GetStagedDiff() (string, error) {
	return Command("diff", "--cached")
}

// WriteTree creates a tree object from the current index
func WriteTree() (string, error) {
	output, err := Command("write-tree")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// CommitTree creates a commit object from a tree
func CommitTree(message string, parent string, tree string) (string, error) {
	args := []string{"commit-tree", "-m", message}
	if parent != "" {
		args = append(args, "-p", parent)
	}
	args = append(args, tree)

	output, err := Command(args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// UpdateRef updates a git reference
func UpdateRef(ref string, commit string) error {
	_, err := Command("update-ref", ref, commit)
	return err
}

// ReadTree reads a tree into the index
func ReadTree(tree string) error {
	_, err := Command("read-tree", tree)
	return err
}

// CheckoutIndex checks out files from the index to the working tree
func CheckoutIndex() error {
	_, err := Command("checkout-index", "-f", "-a")
	return err
}

// UpdateIndex refreshes the index from the working tree
func UpdateIndex() error {
	_, err := Command("update-index", "--refresh")
	return err
}

// ShowRef shows a reference
func ShowRef(ref string) (string, error) {
	output, err := Command("show-ref", ref)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// Show shows a git object
func Show(ref string) (string, error) {
	return Command("show", ref)
}

// ShowFile shows a file at a specific ref
func ShowFile(ref, file string) (string, error) {
	return Command("show", fmt.Sprintf("%s:%s", ref, file))
}

// LsTree lists the contents of a tree
func LsTree(ref string) (string, error) {
	return Command("ls-tree", ref)
}

// Log shows the commit history
func Log(ref string) (string, error) {
	return Command("log", ref)
}

// AddAll stages all changes
func AddAll() error {
	_, err := Command("add", "-A")
	return err
}

// ResetIndex resets the index
func ResetIndex() error {
	_, err := Command("reset", "HEAD")
	return err
}
