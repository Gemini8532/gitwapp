package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
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
