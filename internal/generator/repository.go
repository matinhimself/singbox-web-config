package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	DefaultRepoURL    = "https://github.com/SagerNet/sing-box.git"
	DefaultBranch     = "dev-next"
	DefaultLocalPath  = ".cache/sing-box"
	DefaultRulePath   = "option"
)

// RepositoryManager handles sing-box repository operations
type RepositoryManager struct {
	RepoURL   string
	Branch    string
	LocalPath string
}

// NewRepositoryManager creates a new repository manager with defaults
func NewRepositoryManager() *RepositoryManager {
	return &RepositoryManager{
		RepoURL:   DefaultRepoURL,
		Branch:    DefaultBranch,
		LocalPath: DefaultLocalPath,
	}
}

// WithRepoURL sets a custom repository URL
func (r *RepositoryManager) WithRepoURL(url string) *RepositoryManager {
	r.RepoURL = url
	return r
}

// WithBranch sets a custom branch
func (r *RepositoryManager) WithBranch(branch string) *RepositoryManager {
	r.Branch = branch
	return r
}

// WithLocalPath sets a custom local path
func (r *RepositoryManager) WithLocalPath(path string) *RepositoryManager {
	r.LocalPath = path
	return r
}

// Update clones or updates the sing-box repository
func (r *RepositoryManager) Update() error {
	// Check if directory exists
	if _, err := os.Stat(r.LocalPath); os.IsNotExist(err) {
		// Clone repository
		fmt.Printf("Cloning sing-box repository from %s...\n", r.RepoURL)
		return r.clone()
	}

	// Update existing repository
	fmt.Printf("Updating existing sing-box repository...\n")
	return r.pull()
}

// clone clones the repository
func (r *RepositoryManager) clone() error {
	// Create parent directory
	parentDir := filepath.Dir(r.LocalPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Clone with specific branch
	cmd := exec.Command("git", "clone", "--branch", r.Branch, "--depth", "1", r.RepoURL, r.LocalPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	fmt.Printf("Successfully cloned sing-box repository\n")
	return nil
}

// pull updates the repository
func (r *RepositoryManager) pull() error {
	cmd := exec.Command("git", "-C", r.LocalPath, "pull", "origin", r.Branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull repository: %w", err)
	}

	fmt.Printf("Successfully updated sing-box repository\n")
	return nil
}

// GetRulePath returns the full path to the option directory
func (r *RepositoryManager) GetRulePath() string {
	return filepath.Join(r.LocalPath, DefaultRulePath)
}

// GetCommitHash returns the current commit hash
func (r *RepositoryManager) GetCommitHash() (string, error) {
	cmd := exec.Command("git", "-C", r.LocalPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	return string(output[:7]), nil // Return short hash
}

// GetBranch returns the current branch name
func (r *RepositoryManager) GetBranch() (string, error) {
	cmd := exec.Command("git", "-C", r.LocalPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get branch: %w", err)
	}

	return string(output), nil
}
