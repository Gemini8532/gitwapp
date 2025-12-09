package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Gemini8532/gitwapp/internal/git"
	"github.com/Gemini8532/gitwapp/internal/middleware"
	"github.com/Gemini8532/gitwapp/pkg/models"
	git2 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func setupRepoForTest(t *testing.T) (string, string) {
	tmpDir, err := os.MkdirTemp("", "gitwapp-handler-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	repoPath := filepath.Join(tmpDir, "myrepo")
	r, err := git2.PlainInit(repoPath, false)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to init repo: %v", err)
	}

	w, _ := r.Worktree()
	filename := filepath.Join(repoPath, "README.md")
	os.WriteFile(filename, []byte("Initial"), 0644)
	w.Add("README.md")
	w.Commit("Initial commit", &git2.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})

	return tmpDir, repoPath
}

func addAuth(t *testing.T, req *http.Request) {
	token, err := middleware.GenerateToken("user1", "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
}

func TestHandleRepoStatus(t *testing.T) {
	server, configDir := setupTestServer(t)
	defer os.RemoveAll(configDir)

	tmpDir, repoPath := setupRepoForTest(t)
	defer os.RemoveAll(tmpDir)

	repo := models.Repository{ID: "1", Name: "Test", Path: repoPath}
	server.store.SaveRepositories([]models.Repository{repo})

	// Make a change
	os.WriteFile(filepath.Join(repoPath, "dirty.txt"), []byte("dirty"), 0644)

	req, _ := http.NewRequest("GET", "/api/repos/1/status", nil)
	addAuth(t, req)
	
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %v: %s", rr.Code, rr.Body.String())
	}

	var status git.Status
	if err := json.NewDecoder(rr.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status.Clean {
		t.Error("Expected repo to be dirty")
	}
}

func TestHandleStageAndCommit(t *testing.T) {
	server, configDir := setupTestServer(t)
	defer os.RemoveAll(configDir)

	tmpDir, repoPath := setupRepoForTest(t)
	defer os.RemoveAll(tmpDir)

	repo := models.Repository{ID: "1", Name: "Test", Path: repoPath}
	server.store.SaveRepositories([]models.Repository{repo})

	filename := "newfile.txt"
	os.WriteFile(filepath.Join(repoPath, filename), []byte("content"), 0644)

	// Stage
	stageReq := StageRequest{File: filename}
	body, _ := json.Marshal(stageReq)
	req, _ := http.NewRequest("POST", "/api/repos/1/stage", bytes.NewBuffer(body))
	addAuth(t, req)

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Stage failed: %v %s", rr.Code, rr.Body.String())
	}

	// Commit
	commitReq := CommitRequest{Message: "Added new file"}
	body, _ = json.Marshal(commitReq)
	req, _ = http.NewRequest("POST", "/api/repos/1/commit", bytes.NewBuffer(body))
	addAuth(t, req)

	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Commit failed: %v %s", rr.Code, rr.Body.String())
	}

	// Verify Clean Status
	req, _ = http.NewRequest("GET", "/api/repos/1/status", nil)
	addAuth(t, req)
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	var status git.Status
	json.NewDecoder(rr.Body).Decode(&status)
	if !status.Clean {
		t.Error("Expected repo to be clean after commit")
	}
}
