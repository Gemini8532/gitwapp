package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gemini8532/gitwapp/pkg/models"
)

func TestRunRepoCommand_List(t *testing.T) {
	// Mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos" {
			t.Errorf("Expected path /repos, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method GET, got %s", r.Method)
		}

		repos := []models.Repository{
			{ID: "1", Name: "repo1", Path: "/tmp/repo1"},
		}
		json.NewEncoder(w).Encode(repos)
	}))
	defer ts.Close()

	var out bytes.Buffer
	args := []string{"gitwapp", "repo", "list"}
	err := runRepoCommand(args, ts.URL, &out)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "repo1") {
		t.Errorf("Expected output to contain 'repo1', got %s", output)
	}
}

func TestRunRepoCommand_Add(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos" {
			t.Errorf("Expected path /repos, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"2","name":"newrepo"}`))
	}))
	defer ts.Close()

	var out bytes.Buffer
	args := []string{"gitwapp", "repo", "add", "/tmp/newrepo"}
	err := runRepoCommand(args, ts.URL, &out)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	output := out.String()
	if !strings.Contains(output, "Success") {
		t.Errorf("Expected output to contain 'Success', got %s", output)
	}
}

func TestRunUserCommand_Add(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			t.Errorf("Expected path /users, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	var out bytes.Buffer
	args := []string{"gitwapp", "user", "add", "admin", "secret"}
	err := runUserCommand(args, ts.URL, &out)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestRunRepoCommand_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer ts.Close()

	var out bytes.Buffer
	args := []string{"gitwapp", "repo", "add", "/bad/path"}
	err := runRepoCommand(args, ts.URL, &out)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if !strings.Contains(err.Error(), "Bad Request") {
		t.Errorf("Expected error message to contain 'Bad Request', got %v", err)
	}
}
