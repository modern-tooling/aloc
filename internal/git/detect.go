package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DetectRepo performs a lightweight check for git repository presence
// Returns nil if not a git repo (not an error)
func DetectRepo(root string) (*RepoHint, error) {
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, nil // not a git repo
	}

	hint := &RepoHint{HasGit: true}

	// get first commit timestamp (repo age)
	firstCmd := exec.Command("git", "-C", root, "log", "--reverse", "--format=%aI", "--max-count=1")
	firstOut, err := firstCmd.Output()
	if err == nil && len(firstOut) > 0 {
		firstTime, err := time.Parse(time.RFC3339, strings.TrimSpace(string(firstOut)))
		if err == nil {
			hint.RepoAge = time.Since(firstTime)
		}
	}

	// get last commit timestamp
	lastCmd := exec.Command("git", "-C", root, "log", "-1", "--format=%aI")
	lastOut, err := lastCmd.Output()
	if err == nil && len(lastOut) > 0 {
		lastTime, err := time.Parse(time.RFC3339, strings.TrimSpace(string(lastOut)))
		if err == nil {
			hint.LastCommit = lastTime
			hint.IsActive = time.Since(lastTime) < 7*24*time.Hour
		}
	}

	return hint, nil
}

// IsShallowClone checks if the repo is a shallow clone
func IsShallowClone(root string) bool {
	shallowFile := filepath.Join(root, ".git", "shallow")
	_, err := os.Stat(shallowFile)
	return err == nil
}
