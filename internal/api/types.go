package api

import (
	"time"
)

// User represents a system user
type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

// Repository represents a tracked git repository
type Repository struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	UserID    string    `json:"user_id"`
}

// LoginRequest is the payload for login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the response for login
type LoginResponse struct {
	Token string `json:"token"`
}

// RepoStatus represents the git status of a repository
type RepoStatus struct {
	IsClean    bool   `json:"is_clean"`
	HasChanges bool   `json:"has_changes"` // Dirty
	Ahead      int    `json:"ahead"`
	Behind     int    `json:"behind"`
	CurrentBranch string `json:"current_branch"`
}
