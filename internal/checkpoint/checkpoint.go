package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	
	"agent-ledger/internal/git"
	"agent-ledger/internal/repository"
	"agent-ledger/internal/storage"
)

// Checkpoint represents a checkpoint
type Checkpoint struct {
	ID              string            `json:"id"`
	SessionID       string            `json:"session_id"`
	Ref             string            `json:"ref"`
	Commit          string            `json:"commit"`
	Tree            string            `json:"tree"`
	Timestamp       time.Time          `json:"timestamp"`
	ParentCommit    string            `json:"parent_commit"`
	ParentCheckpoint string           `json:"parent_checkpoint,omitempty"`
	ChangedFiles    []string          `json:"changed_files"`
	AddedFiles      []string          `json:"added_files"`
	DeletedFiles    []string          `json:"deleted_files"`
	ModifiedFiles   []string          `json:"modified_files"`
	GitStatus       GitStatusSnapshot `json:"git_status"`
}

// GitStatusSnapshot captures the git state at checkpoint time
type GitStatusSnapshot struct {
	Branch          string   `json:"branch"`
	Head            string   `json:"head"`
	Dirty           bool     `json:"dirty"`
	StagedFiles     []string `json:"staged_files"`
	UnstagedFiles   []string `json:"unstaged_files"`
	UntrackedFiles  []string `json:"untracked_files"`
}

// Checkpoints holds multiple checkpoints
type Checkpoints struct {
	Checkpoints []Checkpoint `json:"checkpoints"`
}

// Manager manages checkpoints
type Manager struct {
	storage *storage.Storage
}

// NewManager creates a new checkpoint manager
func NewManager(st *storage.Storage) *Manager {
	return &Manager{
		storage: st,
	}
}

// Create creates a checkpoint with index-preserving algorithm
func (m *Manager) Create(sessionID string, repo *repository.Repository) (*Checkpoint, error) {
	// Get current checkpoint count to determine next ID
	count, err := m.getCount(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint count: %w", err)
	}
	
	nextID := fmt.Sprintf("checkpoint-%d", count+1)
	ref := fmt.Sprintf("refs/agents/sessions/%s/%s", sessionID, nextID)
	
	// 1. Save original index state
	originalIndex, err := git.WriteTree()
	if err != nil {
		return nil, fmt.Errorf("failed to save original index: %w", err)
	}
	
	// 2. Backup working tree files that would be affected
	tempDir, err := os.MkdirTemp("", "agent-ledger-backup-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Backup modified and untracked files
	var backedUpFiles []string
	for _, file := range repo.Unstaged {
		if err := backupFile(tempDir, file); err == nil {
			backedUpFiles = append(backedUpFiles, file)
		}
	}
	for _, file := range repo.Untracked {
		if err := backupFile(tempDir, file); err == nil {
			backedUpFiles = append(backedUpFiles, file)
		}
	}
	
	// 3. Stage everything to capture complete working tree
	if err := git.AddAll(); err != nil {
		return nil, fmt.Errorf("failed to stage all changes: %w", err)
	}
	
	// 4. Capture complete tree
	completeTree, err := git.WriteTree()
	if err != nil {
		// Try to restore index before returning error
		git.ReadTree(originalIndex)
		git.CheckoutIndex()
		git.UpdateIndex()
		return nil, fmt.Errorf("failed to capture complete tree: %w", err)
	}
	
	// 5. Create checkpoint commit
	var parent string
	var parentCheckpoint string
	
	// For first checkpoint, parent is HEAD. For subsequent, parent is previous checkpoint.
	if count == 0 {
		parent = repo.Head
	} else {
		prevCheckpointID := fmt.Sprintf("checkpoint-%d", count)
		parentCheckpoint = fmt.Sprintf("refs/agents/sessions/%s/%s", sessionID, prevCheckpointID)
		parent = parentCheckpoint
	}
	
	message := fmt.Sprintf("Agent Ledger checkpoint - session %s - %s", sessionID, nextID)
	commit, err := git.CommitTree(message, parent, completeTree)
	if err != nil {
		// Try to restore index before returning error
		git.ReadTree(originalIndex)
		git.CheckoutIndex()
		git.UpdateIndex()
		return nil, fmt.Errorf("failed to create checkpoint commit: %w", err)
	}
	
	// 6. Restore original index state
	if err := git.ReadTree(originalIndex); err != nil {
		return nil, fmt.Errorf("failed to restore original index: %w", err)
	}
	if err := git.CheckoutIndex(); err != nil {
		return nil, fmt.Errorf("failed to checkout index: %w", err)
	}
	if err := git.UpdateIndex(); err != nil {
		return nil, fmt.Errorf("failed to update index: %w", err)
	}
	
	// 7. Restore working tree files from backup
	for _, file := range backedUpFiles {
		if err := restoreFile(tempDir, file); err != nil {
			// Log warning but continue
			fmt.Printf("Warning: failed to restore file %s: %v\n", file, err)
		}
	}
	
	// 8. Store checkpoint ref
	if err := git.UpdateRef(ref, commit); err != nil {
		return nil, fmt.Errorf("failed to store checkpoint ref: %w", err)
	}
	
	// Create checkpoint metadata
	checkpoint := &Checkpoint{
		ID:              nextID,
		SessionID:       sessionID,
		Ref:             ref,
		Commit:          commit,
		Tree:            completeTree,
		Timestamp:       time.Now().UTC(),
		ParentCommit:    repo.Head,
		ParentCheckpoint: parentCheckpoint,
		ChangedFiles:    repo.Staged,
		AddedFiles:      repo.Untracked,
		DeletedFiles:    []string{}, // Would need more complex git parsing
		ModifiedFiles:   repo.Unstaged,
		GitStatus: GitStatusSnapshot{
			Branch:         repo.Branch,
			Head:           repo.Head,
			Dirty:          repo.Dirty,
			StagedFiles:    repo.Staged,
			UnstagedFiles:  repo.Unstaged,
			UntrackedFiles: repo.Untracked,
		},
	}
	
	// Save checkpoint metadata
	if err := m.addCheckpoint(sessionID, checkpoint); err != nil {
		return nil, fmt.Errorf("failed to save checkpoint metadata: %w", err)
	}
	
	return checkpoint, nil
}

// backupFile backs up a file to a temporary directory
func backupFile(tempDir, file string) error {
	// Make file path absolute relative to repo root
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	
	src, err := os.ReadFile(absFile)
	if err != nil {
		return err
	}
	
	// Create subdirectories in temp dir as needed
	tempFile := filepath.Join(tempDir, file)
	tempDirOfFile := filepath.Dir(tempFile)
	if err := os.MkdirAll(tempDirOfFile, 0755); err != nil {
		return err
	}
	
	return os.WriteFile(tempFile, src, 0644)
}

// restoreFile restores a file from a temporary directory
func restoreFile(tempDir, file string) error {
	tempFile := filepath.Join(tempDir, file)
	src, err := os.ReadFile(tempFile)
	if err != nil {
		return err
	}
	
	// Make file path absolute
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	
	// Create subdirectories as needed
	dirOfFile := filepath.Dir(absFile)
	if err := os.MkdirAll(dirOfFile, 0755); err != nil {
		return err
	}
	
	return os.WriteFile(absFile, src, 0644)
}

// getCount returns the number of checkpoints for a session
func (m *Manager) getCount(sessionID string) (int, error) {
	checkpointsPath := fmt.Sprintf("sessions/%s/checkpoints.json", sessionID)
	
	if !m.storage.FileExists(checkpointsPath) {
		return 0, nil
	}
	
	var checkpoints Checkpoints
	if err := m.storage.ReadJSON(checkpointsPath, &checkpoints); err != nil {
		return 0, err
	}
	
	return len(checkpoints.Checkpoints), nil
}

// addCheckpoint adds a checkpoint to the session's checkpoint list
func (m *Manager) addCheckpoint(sessionID string, checkpoint *Checkpoint) error {
	checkpointsPath := fmt.Sprintf("sessions/%s/checkpoints.json", sessionID)
	
	var checkpoints Checkpoints
	if m.storage.FileExists(checkpointsPath) {
		if err := m.storage.ReadJSON(checkpointsPath, &checkpoints); err != nil {
			return err
		}
	}
	
	checkpoints.Checkpoints = append(checkpoints.Checkpoints, *checkpoint)
	
	return m.storage.WriteJSON(checkpointsPath, checkpoints)
}

// List lists all checkpoints for a session
func (m *Manager) List(sessionID string) ([]Checkpoint, error) {
	checkpointsPath := fmt.Sprintf("sessions/%s/checkpoints.json", sessionID)
	
	if !m.storage.FileExists(checkpointsPath) {
		return []Checkpoint{}, nil
	}
	
	var checkpoints Checkpoints
	if err := m.storage.ReadJSON(checkpointsPath, &checkpoints); err != nil {
		return nil, err
	}
	
	return checkpoints.Checkpoints, nil
}

// Get retrieves a specific checkpoint
func (m *Manager) Get(sessionID, checkpointID string) (*Checkpoint, error) {
	checkpoints, err := m.List(sessionID)
	if err != nil {
		return nil, err
	}
	
	for _, cp := range checkpoints {
		if cp.ID == checkpointID {
			return &cp, nil
		}
	}
	
	return nil, fmt.Errorf("checkpoint not found: %s", checkpointID)
}

// GetFileAtCheckpoint retrieves a file's content at a specific checkpoint
func (m *Manager) GetFileAtCheckpoint(ref, file string) (string, error) {
	return git.ShowFile(ref, file)
}

// DiffCheckpoints shows the diff between two checkpoints
func (m *Manager) DiffCheckpoints(ref1, ref2 string) (string, error) {
	return git.Command("diff", ref1, ref2)
}
