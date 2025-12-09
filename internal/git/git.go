package git

import (
	"github.com/go-git/go-git/v5"
)

// Status represents the high-level status of a repository
type Status struct {
	Clean       bool
	Ahead       int
	Behind      int
	Branch      string
	Worktree    git.Status
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

	// Calculate ahead/behind
	// This is a simplified check. Real implementation needs to compare with upstream.
	ahead, behind := 0, 0
	
	// Check for upstream tracking
	// refConfig, err := r.Config()
	// if err == nil {
	// 	// Logic to find tracking branch and compare
	// }

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
	
	// For MVP, we assume using SSH keys from the system or standard config
	return r.Push(&git.PushOptions{})
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

	return w.Pull(&git.PullOptions{
		RemoteName: "origin",
	})
}

// Helper to check if a valid repo exists at path
func IsRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}
