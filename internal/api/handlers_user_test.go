package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Gemini8532/gitwapp/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

func TestHandleAddUser(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer os.RemoveAll(tmpDir)

	reqBody := AddUserRequest{
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/internal/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	var user models.User
	if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", user.Username)
	}
	if user.PasswordHash != "" {
		t.Error("Expected password hash to be empty in response")
	}

	// Verify in store
	users, _ := server.store.LoadUsers()
	if len(users) != 1 {
		t.Fatalf("Expected 1 user in store, got %d", len(users))
	}
	if err := bcrypt.CompareHashAndPassword([]byte(users[0].PasswordHash), []byte("password123")); err != nil {
		t.Errorf("Password hash verification failed")
	}
}

func TestHandleListUsers(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer os.RemoveAll(tmpDir)

	server.store.SaveUsers([]models.User{
		{ID: "1", Username: "u1", PasswordHash: "hash1"},
		{ID: "2", Username: "u2", PasswordHash: "hash2"},
	})

	req, _ := http.NewRequest("GET", "/internal/api/users", nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var users []models.User
	if err := json.NewDecoder(rr.Body).Decode(&users); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	if users[0].PasswordHash != "" {
		t.Error("Expected password hash to be scrubbed")
	}
}

func TestHandleRemoveUser(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer os.RemoveAll(tmpDir)

	server.store.SaveUsers([]models.User{
		{ID: "1", Username: "u1", PasswordHash: "hash1"},
	})

	req, _ := http.NewRequest("DELETE", "/internal/api/users/1", nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}

	users, _ := server.store.LoadUsers()
	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}
