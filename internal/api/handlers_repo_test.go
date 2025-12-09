package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Gemini8532/gitwapp/internal/config"
	"github.com/Gemini8532/gitwapp/pkg/models"
)

func setupTestServer(t *testing.T) (*Server, string) {
	tmpDir, err := os.MkdirTemp("", "gitwapp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	store, err := config.NewStoreWithDir(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create store: %v", err)
	}

	server := NewServer(store)
	return server, tmpDir
}

func TestHandleAddRepo(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer os.RemoveAll(tmpDir)

	// Create a dummy repo directory
	repoPath := filepath.Join(tmpDir, "myrepo")
	if err := os.Mkdir(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create dummy repo dir: %v", err)
	}

	reqBody := AddRepoRequest{
		Path: repoPath,
		Name: "My Repo",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/internal/api/repos", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	var repo models.Repository
	if err := json.NewDecoder(rr.Body).Decode(&repo); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if repo.Path != repoPath {
		t.Errorf("Expected path %s, got %s", repoPath, repo.Path)
	}
	if repo.Name != "My Repo" {
		t.Errorf("Expected name 'My Repo', got %s", repo.Name)
	}
	if repo.ID == "" {
		t.Error("Expected ID to be set")
	}
}

func TestHandleListRepos(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer os.RemoveAll(tmpDir)

	// Add a repo directly to store
	repo := models.Repository{
		ID:   "123",
		Name: "Test Repo",
		Path: "/tmp/test",
	}
	server.store.SaveRepositories([]models.Repository{repo})

	req, _ := http.NewRequest("GET", "/internal/api/repos", nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var repos []models.Repository
	if err := json.NewDecoder(rr.Body).Decode(&repos); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(repos))
	}
	if repos[0].ID != "123" {
		t.Errorf("Expected repo ID '123', got %s", repos[0].ID)
	}
}

func TestHandleRemoveRepo(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer os.RemoveAll(tmpDir)

	repo := models.Repository{
		ID:   "123",
		Name: "Test Repo",
		Path: "/tmp/test",
	}
	server.store.SaveRepositories([]models.Repository{repo})

	req, _ := http.NewRequest("DELETE", "/internal/api/repos/123", nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}

	// Verify it's gone
	repos, _ := server.store.LoadRepositories()
	if len(repos) != 0 {
		t.Errorf("Expected 0 repos, got %d", len(repos))
	}
}
