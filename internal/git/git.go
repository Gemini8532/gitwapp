package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// Status represents the high-level status of a repository
type Status struct {
	Clean    bool
	Ahead    int
	Behind   int
	Branch   string
	Worktree git.Status
}

type Client struct {
	// wrapper if we need to store config
}

func Open(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

func GetStatus(path string) (*Status, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := w.Status()
	if err != nil {
		return nil, err
	}

	head, err := r.Head()
	if err != nil {
		return nil, err
	}

	branchName := head.Name().Short()

	// Calculate ahead/behind by comparing with remote tracking branch
	ahead, behind := 0, 0

	// Get the remote tracking branch
	remoteBranchName := "refs/remotes/origin/" + branchName
	remoteRef, err := r.Reference(plumbing.ReferenceName(remoteBranchName), true)
	if err == nil {
		// Both local and remote refs exist, compare them
		localCommit, err := r.CommitObject(head.Hash())
		if err == nil {
			remoteCommit, err := r.CommitObject(remoteRef.Hash())
			if err == nil {
				// Count commits ahead (local commits not in remote)
				ahead, _ = countCommitsBetween(r, localCommit.Hash, remoteCommit.Hash)
				// Count commits behind (remote commits not in local)
				behind, _ = countCommitsBetween(r, remoteCommit.Hash, localCommit.Hash)
			}
		}
	}
	// If remote branch doesn't exist, ahead/behind stay at 0

	return &Status{
		Clean:    status.IsClean(),
		Ahead:    ahead,
		Behind:   behind,
		Branch:   branchName,
		Worktree: status,
	}, nil
}

func StageFile(path string, file string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(file)
	return err
}

func UnstageFile(path string, file string) error {
	// go-git doesn't have a clean API for unstaging a single file
	// Use git command directly for reliability
	cmd := exec.Command("git", "reset", "HEAD", "--", file)
	cmd.Dir = path
	return cmd.Run()
}

func StageAll(path string) error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = path
	return cmd.Run()
}

func UnstageAll(path string) error {
	cmd := exec.Command("git", "reset")
	cmd.Dir = path
	return cmd.Run()
}

func GetFileDiff(path string, file string) (string, error) {
	// Get unified diff for the file
	cmd := exec.Command("git", "diff", "HEAD", "--", file)
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func Commit(path string, msg string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Commit(msg, &git.CommitOptions{})
	return err
}

func Push(path string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	auth, err := getAuth(path)
	if err != nil {
		return fmt.Errorf("failed to get auth: %w", err)
	}

	return r.Push(&git.PushOptions{
		Auth: auth,
	})
}

func Pull(path string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	auth, err := getAuth(path)
	if err != nil {
		return fmt.Errorf("failed to get auth: %w", err)
	}

	return w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
}

// getSSHAuth attempts to get SSH authentication using ssh-agent or key files
func getSSHAuth() (transport.AuthMethod, error) {
	// Try ssh-agent first (most common and secure)
	auth, err := ssh.NewSSHAgentAuth("git")
	if err == nil {
		return auth, nil
	}

	// Fallback: try common SSH key locations
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	keyPaths := []string{
		filepath.Join(homeDir, ".ssh", "id_ed25519"),
		filepath.Join(homeDir, ".ssh", "id_rsa"),
		filepath.Join(homeDir, ".ssh", "id_ecdsa"),
	}

	for _, keyPath := range keyPaths {
		auth, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
		if err == nil {
			return auth, nil
		}
	}

	return nil, fmt.Errorf("no SSH authentication method available")
}

// getAuth gets appropriate authentication based on remote URL
func getAuth(repoPath string) (transport.AuthMethod, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	// Get the remote URL to determine auth type
	remote, err := r.Remote("origin")
	if err != nil {
		return nil, err
	}

	if len(remote.Config().URLs) == 0 {
		return nil, fmt.Errorf("no remote URL configured")
	}

	remoteURL := remote.Config().URLs[0]

	// Check if it's SSH or HTTPS
	if strings.HasPrefix(remoteURL, "git@") || strings.HasPrefix(remoteURL, "ssh://") {
		return getSSHAuth()
	}

	// For HTTPS, use git credential helper
	return getHTTPSAuth(remoteURL)
}

// getHTTPSAuth gets credentials from git credential helper
func getHTTPSAuth(remoteURL string) (transport.AuthMethod, error) {
	// Use git credential fill to get credentials
	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("protocol=https\nhost=github.com\npath=%s\n\n", remoteURL))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// Parse the output
	var username, password string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "username=") {
			username = strings.TrimPrefix(line, "username=")
		} else if strings.HasPrefix(line, "password=") {
			password = strings.TrimPrefix(line, "password=")
		}
	}

	if username == "" || password == "" {
		return nil, fmt.Errorf("no credentials found")
	}

	return &http.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}

// countCommitsBetween counts how many commits are in 'from' that are not in 'to'
func countCommitsBetween(r *git.Repository, from, to plumbing.Hash) (int, error) {
	if from == to {
		return 0, nil
	}

	// Get commit iterator starting from 'from'
	commitIter, err := r.Log(&git.LogOptions{From: from})
	if err != nil {
		return 0, err
	}
	defer commitIter.Close()

	count := 0
	err = commitIter.ForEach(func(c *object.Commit) error {
		if c.Hash == to {
			// Reached the 'to' commit, stop counting
			return storer.ErrStop
		}
		count++
		return nil
	})

	if err != nil && err != storer.ErrStop {
		return 0, err
	}

	return count, nil
}

// Helper to check if a valid repo exists at path
func IsRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}
