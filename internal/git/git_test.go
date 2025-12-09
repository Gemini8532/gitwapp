package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func setupTestRepo(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "gitwapp-git-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize a new git repo
	r, err := git.PlainInit(tmpDir, false)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to init git repo: %v", err)
	}

	w, _ := r.Worktree()
	
	// Create a file and commit it
	filename := filepath.Join(tmpDir, "README.md")
	os.WriteFile(filename, []byte("Hello"), 0644)
	w.Add("README.md")
	w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})

	return tmpDir
}

func setupTestRepoWithRemote(t *testing.T) (string, string) {
	// 1. Create bare remote directory
	remoteDir, err := os.MkdirTemp("", "gitwapp-remote-test")
	if err != nil {
		t.Fatalf("Failed to create remote dir: %v", err)
	}
	
	// Initialize bare repo
	_, err = git.PlainInit(remoteDir, true) 
	if err != nil {
		os.RemoveAll(remoteDir)
		t.Fatalf("Failed to init bare repo: %v", err)
	}

	// 2. Create a temporary local repo to seed the remote
	seedDir, err := os.MkdirTemp("", "gitwapp-seed-test")
	if err != nil {
		os.RemoveAll(remoteDir)
		t.Fatalf("Failed to create seed dir: %v", err)
	}
	defer os.RemoveAll(seedDir)

	rSeed, err := git.PlainInit(seedDir, false)
	if err != nil {
		t.Fatalf("Failed to init seed repo: %v", err)
	}

	wSeed, _ := rSeed.Worktree()
	filename := filepath.Join(seedDir, "README.md")
	os.WriteFile(filename, []byte("Hello"), 0644)
	wSeed.Add("README.md")
	wSeed.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})

	// Add remote and push
	_, err = rSeed.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteDir},
	})
	if err != nil {
		t.Fatalf("Failed to create remote: %v", err)
	}

	err = rSeed.Push(&git.PushOptions{
		RemoteName: "origin",
	})
	if err != nil {
		t.Fatalf("Failed to push seed: %v", err)
	}

	// 3. Create the actual local test repo by cloning the now-populated remote
	localDir, err := os.MkdirTemp("", "gitwapp-local-test")
	if err != nil {
		os.RemoveAll(remoteDir)
		t.Fatalf("Failed to create local dir: %v", err)
	}

	_, err = git.PlainClone(localDir, false, &git.CloneOptions{
		URL: remoteDir,
	})
	if err != nil {
		os.RemoveAll(remoteDir)
		os.RemoveAll(localDir)
		t.Fatalf("Failed to clone repo: %v", err)
	}

	return localDir, remoteDir
}

func TestIsRepo(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer os.RemoveAll(repoPath)

	if !IsRepo(repoPath) {
		t.Errorf("Expected IsRepo to return true for %s", repoPath)
	}

	if IsRepo("/invalid/path") {
		t.Error("Expected IsRepo to return false for invalid path")
	}
}

func TestGetStatus(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer os.RemoveAll(repoPath)

	// Test Clean Status
	status, err := GetStatus(repoPath)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if !status.Clean {
		t.Error("Expected repo to be clean")
	}

	// Make it dirty
	os.WriteFile(filepath.Join(repoPath, "newfile.txt"), []byte("dirty"), 0644)

	status, err = GetStatus(repoPath)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if status.Clean {
		t.Error("Expected repo to be dirty")
	}
}

func TestStageAndCommit(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer os.RemoveAll(repoPath)

	filename := "test.txt"
	fullPath := filepath.Join(repoPath, filename)
	os.WriteFile(fullPath, []byte("content"), 0644)

	// Stage
	if err := StageFile(repoPath, filename); err != nil {
		t.Fatalf("StageFile failed: %v", err)
	}

	// Verify staged
	status, _ := GetStatus(repoPath)
	s := status.Worktree.File(filename)
	if s.Staging != git.Added {
		t.Errorf("Expected file to be staged as Added, got %v", s.Staging)
	}

	// Commit
	if err := Commit(repoPath, "Add test file"); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify clean
	status, _ = GetStatus(repoPath)
	if !status.Clean {
		t.Error("Expected repo to be clean after commit")
	}
}

func TestPush(t *testing.T) {
	local, remote := setupTestRepoWithRemote(t)
	defer os.RemoveAll(local)
	defer os.RemoveAll(remote)

	// Make change
	filename := filepath.Join(local, "push.txt")
	os.WriteFile(filename, []byte("push"), 0644)
	
	r, _ := git.PlainOpen(local)
	w, _ := r.Worktree()
	w.Add("push.txt")
	w.Commit("Push commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})

	if err := Push(local); err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Verify remote has the commit (by cloning it elsewhere or checking ref)
	rRemote, _ := git.PlainOpen(remote)
	ref, err := rRemote.Head()
	if err != nil {
		t.Fatal("Failed to get remote HEAD")
	}
	
	// Local HEAD should match Remote HEAD
	refLocal, _ := r.Head()
	if ref.Hash() != refLocal.Hash() {
		t.Errorf("Remote HEAD %v does not match Local HEAD %v", ref.Hash(), refLocal.Hash())
	}
}

func TestPull(t *testing.T) {
	// Setup: Remote and User A (local)
	localA, remote := setupTestRepoWithRemote(t)
	defer os.RemoveAll(localA)
	defer os.RemoveAll(remote)

	// User B clones the same remote
	localB, err := os.MkdirTemp("", "gitwapp-local-b")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(localB)

	_, err = git.PlainClone(localB, false, &git.CloneOptions{
		URL: remote,
	})
	if err != nil {
		t.Fatal(err)
	}

	// User A pushes a change
	filename := filepath.Join(localA, "pull.txt")
	os.WriteFile(filename, []byte("pull"), 0644)
	rA, _ := git.PlainOpen(localA)
	wA, _ := rA.Worktree()
	wA.Add("pull.txt")
	wA.Commit("Pull commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	rA.Push(&git.PushOptions{})

	// User B pulls
	if err := Pull(localB); err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	// Verify file exists in User B
	if _, err := os.Stat(filepath.Join(localB, "pull.txt")); os.IsNotExist(err) {
		t.Error("Expected pull.txt to exist in localB after pull")
	}
}
