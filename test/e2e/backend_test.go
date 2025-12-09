package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Gemini8532/gitwapp/internal/api"
	"github.com/Gemini8532/gitwapp/internal/config"
	"github.com/Gemini8532/gitwapp/pkg/models"
	git2 "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type TestContext struct {
	Server    *httptest.Server
	ConfigDir string
	Client    *http.Client
	Token     string
}

func setupE2E(t *testing.T) *TestContext {
	configDir, err := os.MkdirTemp("", "gitwapp-e2e-config")
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	store, err := config.NewStoreWithDir(configDir)
	if err != nil {
		os.RemoveAll(configDir)
		t.Fatalf("Failed to create store: %v", err)
	}

	apiServer := api.NewServer(store)
	ts := httptest.NewServer(apiServer.Handler())

	return &TestContext{
		Server:    ts,
		ConfigDir: configDir,
		Client:    ts.Client(),
	}
}

func (tc *TestContext) Teardown() {
	tc.Server.Close()
	os.RemoveAll(tc.ConfigDir)
}

func (tc *TestContext) createUser(t *testing.T, username, password string) {
	reqBody := api.AddUserRequest{Username: username, Password: password}
	body, _ := json.Marshal(reqBody)
	resp, err := tc.Client.Post(tc.Server.URL+"/internal/api/users", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create user failed: %v", resp.Status)
	}
}

func (tc *TestContext) login(t *testing.T, username, password string) {
	reqBody := api.LoginRequest{Username: username, Password: password}
	body, _ := json.Marshal(reqBody)
	resp, err := tc.Client.Post(tc.Server.URL+"/api/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login returned status: %v", resp.Status)
	}

	var loginResp api.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}
	tc.Token = loginResp.Token
}

func (tc *TestContext) authenticatedRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, tc.Server.URL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tc.Token)
	req.Header.Set("Content-Type", "application/json")

	return tc.Client.Do(req)
}

func setupGitRepo(t *testing.T) (string, string, func()) {
	// Create bare remote
	remoteDir, err := os.MkdirTemp("", "gitwapp-e2e-remote")
	if err != nil {
		t.Fatal(err)
	}
	_, err = git2.PlainInit(remoteDir, true) // Bare repo
	if err != nil {
		os.RemoveAll(remoteDir)
		t.Fatal(err)
	}

	// Create local repo
	localDir, err := os.MkdirTemp("", "gitwapp-e2e-local")
	if err != nil {
		os.RemoveAll(remoteDir)
		t.Fatal(err)
	}
	
	// Clone from remote (initially empty)
	r, err := git2.PlainClone(localDir, false, &git2.CloneOptions{
		URL: remoteDir,
	})
	if err != nil {
		// Clone of an empty bare repo can fail, so we'll init locally and add remote
		// If clone fails, try init locally and add remote
		r, err = git2.PlainInit(localDir, false)
		if err != nil {
			os.RemoveAll(remoteDir)
			os.RemoveAll(localDir)
			t.Fatal(err)
		}
		_, err = r.CreateRemote(&gitconfig.RemoteConfig{
			Name: git2.DefaultRemoteName,
			URLs: []string{remoteDir},
		})
		if err != nil {
			os.RemoveAll(remoteDir)
			os.RemoveAll(localDir)
			t.Fatal(err)
		}
	}

	w, _ := r.Worktree()
	os.WriteFile(filepath.Join(localDir, "README.md"), []byte("# Test"), 0644)
	w.Add("README.md")
	w.Commit("Init", &git2.CommitOptions{
		Author: &object.Signature{Name: "Tester", Email: "test@test.com", When: time.Now()},
	})

	// Push initial commit to remote
	err = r.Push(&git2.PushOptions{})
	if err != nil {
		os.RemoveAll(remoteDir)
		os.RemoveAll(localDir)
		t.Fatal(err)
	}

	return localDir, remoteDir, func() {
		os.RemoveAll(localDir)
		os.RemoveAll(remoteDir)
	}
}

func TestE2E_FullFlow(t *testing.T) {
	tc := setupE2E(t)
	defer tc.Teardown()

	// 1. Setup User via Internal API
	tc.createUser(t, "admin", "secret123")

	// 2. Login via Public API
	tc.login(t, "admin", "secret123")
		if tc.Token == "" {
			t.Fatal("Token should be set")
		}
	
		// 3. Setup a Git Repo on disk, connected to a remote
		repoPath, remotePath, cleanupRepo := setupGitRepo(t)
		defer cleanupRepo()
	
		// 4. Add Repo via Internal API
		addRepoBody := api.AddRepoRequest{Path: repoPath, Name: "My E2E Repo"}
		jsonBody, _ := json.Marshal(addRepoBody)
		resp, err := tc.Client.Post(tc.Server.URL+"/internal/api/repos", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Add repo failed: %v", resp.Status)
		}
		var repo models.Repository
		json.NewDecoder(resp.Body).Decode(&repo)
		repoID := repo.ID
	
		// 5. Verify Repo List via Public API (Authenticated)
		resp, err = tc.authenticatedRequest("GET", "/api/repos", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Get repos failed: %v", resp.Status)
		}
	
		// 6. Check Status (Clean)
		resp, err = tc.authenticatedRequest("GET", fmt.Sprintf("/api/repos/%s/status", repoID), nil)
		if err != nil {
			t.Fatal(err)
		}
		// We need to decode the status structure. Since it's internal/git, we can't import `git` package easily if it has common name.
		// But we imported `internal/git` as `git` is likely shadowed? 
		// No, we haven't imported internal/git in this file yet for types.
		// The response is JSON, we can check basic string content or decode into map.
		bodyBytes, _ := io.ReadAll(resp.Body)
		if !bytes.Contains(bodyBytes, []byte(`"Clean":true`)) { // Rough check
			t.Errorf("Expected Clean:true, got %s", string(bodyBytes))
		}
	
		// 7. Make Changes and Commit via API
		os.WriteFile(filepath.Join(repoPath, "change.txt"), []byte("changed"), 0644)
	
		// Check status again (Dirty)
		resp, err = tc.authenticatedRequest("GET", fmt.Sprintf("/api/repos/%s/status", repoID), nil)
		bodyBytes, _ = io.ReadAll(resp.Body)
		if !bytes.Contains(bodyBytes, []byte(`"Clean":false`)) {
			t.Errorf("Expected Clean:false after change, got %s", string(bodyBytes))
		}
	
		// Stage
		stageReq := api.StageRequest{File: "change.txt"}
		resp, err = tc.authenticatedRequest("POST", fmt.Sprintf("/api/repos/%s/stage", repoID), stageReq)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Stage failed: %v", resp.Status)
		}
	
		// Commit
		commitReq := api.CommitRequest{Message: "E2E Commit"}
		resp, err = tc.authenticatedRequest("POST", fmt.Sprintf("/api/repos/%s/commit", repoID), commitReq)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Commit failed: %v", resp.Status)
		}
	
		// Check status again (Clean)
		resp, err = tc.authenticatedRequest("GET", fmt.Sprintf("/api/repos/%s/status", repoID), nil)
		bodyBytes, _ = io.ReadAll(resp.Body)
		if !bytes.Contains(bodyBytes, []byte(`"Clean":true`)) {
			t.Errorf("Expected Clean:true after commit, got %s", string(bodyBytes))
		}
	
		// 8. Push to remote
		resp, err = tc.authenticatedRequest("POST", fmt.Sprintf("/api/repos/%s/push", repoID), nil)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Push failed: %v", resp.Status)
		}
		
		// Verify remote has the commit - by inspecting remote directly
		rRemote, _ := git2.PlainOpen(remotePath)
		refRemote, _ := rRemote.Head()
		rLocal, _ := git2.PlainOpen(repoPath)
		refLocal, _ := rLocal.Head()
		if refRemote.Hash() != refLocal.Hash() {
			t.Errorf("Remote HEAD %v does not match Local HEAD %v after push", refRemote.Hash(), refLocal.Hash())
		}
	
		// 9. Simulate remote change and Pull
		// Create another local clone to make a change and push it to remote
		otherLocalDir, err := os.MkdirTemp("", "gitwapp-e2e-other-local")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(otherLocalDir)
		
		rOther, _ := git2.PlainClone(otherLocalDir, false, &git2.CloneOptions{
			URL: remotePath,
		})
		wOther, _ := rOther.Worktree()
		os.WriteFile(filepath.Join(otherLocalDir, "remote_change.txt"), []byte("remote content"), 0644)
		wOther.Add("remote_change.txt")
		wOther.Commit("Remote commit", &git2.CommitOptions{
			Author: &object.Signature{Name: "Remote", Email: "remote@test.com", When: time.Now()},
		})
		rOther.Push(&git2.PushOptions{})
	
		// Pull via API
		resp, err = tc.authenticatedRequest("POST", fmt.Sprintf("/api/repos/%s/pull", repoID), nil)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Pull failed: %v", resp.Status)
		}
	
		// Verify local repo has the remote change
		if _, err := os.Stat(filepath.Join(repoPath, "remote_change.txt")); os.IsNotExist(err) {
			t.Error("Expected remote_change.txt to exist in local repo after pull")
		}
	}
	