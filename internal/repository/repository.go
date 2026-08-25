package repository

import (
	"fmt"
	"strings"
	"agent-ledger/internal/git"
)

// Repository represents a git repository
type Repository struct {
	Root     string
	Branch   string
	Head     string
	Remotes  map[string]string
	Dirty    bool
	Staged   []string
	Unstaged []string
	Untracked []string
}

// Detect detects the current git repository
func Detect() (*Repository, error) {
	isRepo, err := git.IsRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to check if repository: %w", err)
	}
	
	if !isRepo {
		return nil, fmt.Errorf("not a git repository")
	}
	
	root, err := git.GetRepositoryRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}
	
	branch, err := git.GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	
	head, err := git.GetHeadCommit()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}
	
	remotes, err := git.GetRemotes()
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
	}
	
	dirty, err := git.IsDirty()
	if err != nil {
		return nil, fmt.Errorf("failed to check dirty state: %w", err)
	}
	
	repo := &Repository{
		Root:    root,
		Branch:  branch,
		Head:    head,
		Remotes: remotes,
		Dirty:   dirty,
	}
	
	// Parse status for detailed file information
	if dirty {
		repo.parseStatus()
	}
	
	return repo, nil
}

// parseStatus parses git status to extract staged, unstaged, and untracked files
func (r *Repository) parseStatus() {
	status, err := git.GetStatus()
	if err != nil {
		return
	}
	
	lines := strings.Split(status, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if len(line) < 3 {
			continue
		}
		
		statusCode := line[:2]
		filename := strings.TrimSpace(line[3:])
		
		switch statusCode[0] {
		case 'M', 'A', 'D', 'R', 'C':
			r.Staged = append(r.Staged, filename)
		}
		
		switch statusCode[1] {
		case 'M', 'D':
			r.Unstaged = append(r.Unstaged, filename)
		}
		
		if statusCode == "??" {
			r.Untracked = append(r.Untracked, filename)
		}
	}
}

// MustBeInRepository checks if we're in a git repository and returns an error if not
func MustBeInRepository() error {
	isRepo, err := git.IsRepository()
	if err != nil {
		return fmt.Errorf("failed to check if repository: %w", err)
	}
	
	if !isRepo {
		return fmt.Errorf("not a git repository - agent-ledger requires a git repository to work")
	}
	
	return nil
}

// GetRepositoryRoot returns the repository root path
func GetRepositoryRoot() (string, error) {
	return git.GetRepositoryRoot()
}
