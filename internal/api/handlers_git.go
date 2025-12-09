package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gemini8532/gitwapp/internal/git"
	"github.com/Gemini8532/gitwapp/pkg/models"
	"github.com/gorilla/mux"
)

// handleGetRepos handles requests to retrieve all repositories. It is part of the public API.
func (s *Server) handleGetRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slog.InfoContext(ctx, "Getting all repositories")
	repos, err := s.store.LoadRepositories()
	if err != nil {
		slog.ErrorContext(ctx, "Get repos failed - unable to load repositories", "error", err)
		http.Error(w, "Failed to load repositories", http.StatusInternalServerError)
		return
	}

	// TODO: Maybe enrich with simple status (clean/dirty) if performance allows
	// For now, just return the list
	slog.InfoContext(ctx, "Repositories retrieved successfully", "count", len(repos))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repos)
}

// handleRepoStatus handles requests for the detailed status of a single repository.
// It is part of the public API.
func (s *Server) handleRepoStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	slog.InfoContext(ctx, "Getting repository status", "id", id)

	repo, err := s.getRepoByID(id)
	if err != nil {
		slog.WarnContext(ctx, "Get repo status failed - repository not found", "id", id)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	status, err := git.GetStatus(repo.Path)
	if err != nil {
		slog.ErrorContext(ctx, "Get repo status failed - unable to get git status", "id", id, "path", repo.Path, "error", err)
		http.Error(w, "Failed to get git status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Repository status retrieved successfully", "id", id, "path", repo.Path)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// StageRequest represents the request body for staging or unstaging a file.
type StageRequest struct {
	File string `json:"file"`
}

// handleStage handles requests to stage a single file in a repository.
func (s *Server) handleStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	var req StageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "Failed to decode stage request", "id", id, "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "Staging file", "id", id, "file", req.File)

	repo, err := s.getRepoByID(id)
	if err != nil {
		slog.WarnContext(ctx, "Stage file failed - repository not found", "id", id, "file", req.File)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.StageFile(repo.Path, req.File); err != nil {
		slog.ErrorContext(ctx, "Stage file failed", "id", id, "file", req.File, "path", repo.Path, "error", err)
		http.Error(w, "Failed to stage file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "File staged successfully", "id", id, "file", req.File, "path", repo.Path)
	w.WriteHeader(http.StatusOK)
}

// handleUnstage handles requests to unstage a single file in a repository.
func (s *Server) handleUnstage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	var req StageRequest // Reuse same request struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "Failed to decode unstage request", "id", id, "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "Unstaging file", "id", id, "file", req.File)

	repo, err := s.getRepoByID(id)
	if err != nil {
		slog.WarnContext(ctx, "Unstage file failed - repository not found", "id", id, "file", req.File)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.UnstageFile(repo.Path, req.File); err != nil {
		slog.ErrorContext(ctx, "Unstage file failed", "id", id, "file", req.File, "path", repo.Path, "error", err)
		http.Error(w, "Failed to unstage file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "File unstaged successfully", "id", id, "file", req.File, "path", repo.Path)
	w.WriteHeader(http.StatusOK)
}

// handleStageAll handles requests to stage all modified and new files in a repository.
func (s *Server) handleStageAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	slog.InfoContext(ctx, "Staging all files", "id", id)

	repo, err := s.getRepoByID(id)
	if err != nil {
		slog.WarnContext(ctx, "Stage all failed - repository not found", "id", id)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.StageAll(repo.Path); err != nil {
		slog.ErrorContext(ctx, "Stage all failed", "id", id, "path", repo.Path, "error", err)
		http.Error(w, "Failed to stage files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "All files staged successfully", "id", id, "path", repo.Path)
	w.WriteHeader(http.StatusOK)
}

// handleUnstageAll handles requests to unstage all staged files in a repository.
func (s *Server) handleUnstageAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	slog.InfoContext(ctx, "Unstaging all files", "id", id)

	repo, err := s.getRepoByID(id)
	if err != nil {
		slog.WarnContext(ctx, "Unstage all failed - repository not found", "id", id)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.UnstageAll(repo.Path); err != nil {
		slog.ErrorContext(ctx, "Unstage all failed", "id", id, "path", repo.Path, "error", err)
		http.Error(w, "Failed to unstage files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "All files unstaged successfully", "id", id, "path", repo.Path)
	w.WriteHeader(http.StatusOK)
}

// handleGetFile handles requests to retrieve the content of a specific file from a repository's working directory.
func (s *Server) handleGetFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]
	file := r.URL.Query().Get("file")

	if file == "" {
		http.Error(w, "File parameter is required", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "Getting file content", "id", id, "file", file)

	repo, err := s.getRepoByID(id)
	if err != nil {
		slog.WarnContext(ctx, "Get file failed - repository not found", "id", id)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	// Security check: ensure path is within repo
	targetPath := filepath.Clean(filepath.Join(repo.Path, file))
	if !strings.HasPrefix(targetPath, filepath.Clean(repo.Path)) {
		slog.WarnContext(ctx, "Get file failed - path traversal attempt", "id", id, "file", file)
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		slog.ErrorContext(ctx, "Get file failed - read error", "id", id, "file", file, "error", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Detect content type or default to plain text
	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

// handleGetDiff handles requests to get the diff of a specific file in a repository.
func (s *Server) handleGetDiff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]
	file := r.URL.Query().Get("file")

	if file == "" {
		http.Error(w, "File parameter is required", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "Getting file diff", "id", id, "file", file)

	repo, err := s.getRepoByID(id)
	if err != nil {
		slog.WarnContext(ctx, "Get diff failed - repository not found", "id", id)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	diff, err := git.GetFileDiff(repo.Path, file)
	if err != nil {
		slog.ErrorContext(ctx, "Get diff failed", "id", id, "file", file, "path", repo.Path, "error", err)
		http.Error(w, "Failed to get diff: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(diff))
}

// CommitRequest represents the request body for committing changes.
type CommitRequest struct {
	Message string `json:"message"`
}

// handleCommit handles requests to commit staged changes in a repository.
func (s *Server) handleCommit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	var req CommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "Failed to decode commit request", "id", id, "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		slog.WarnContext(ctx, "Commit failed - commit message required", "id", id)
		http.Error(w, "Commit message required", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "Committing changes", "id", id, "message", req.Message)

	repo, err := s.getRepoByID(id)
	if err != nil {
		slog.WarnContext(ctx, "Commit failed - repository not found", "id", id)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.Commit(repo.Path, req.Message); err != nil {
		slog.ErrorContext(ctx, "Commit failed", "id", id, "path", repo.Path, "error", err)
		http.Error(w, "Failed to commit: "+err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Changes committed successfully", "id", id, "path", repo.Path)
	w.WriteHeader(http.StatusOK)
}

// handlePush handles requests to push committed changes to a remote repository.
func (s *Server) handlePush(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	slog.InfoContext(ctx, "Pushing changes", "id", id)

	repo, err := s.getRepoByID(id)
	if err != nil {
		slog.WarnContext(ctx, "Push failed - repository not found", "id", id)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.Push(repo.Path); err != nil {
		slog.ErrorContext(ctx, "Push failed", "id", id, "path", repo.Path, "error", err)
		http.Error(w, "Failed to push: "+err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Changes pushed successfully", "id", id, "path", repo.Path)
	w.WriteHeader(http.StatusOK)
}

// handlePull handles requests to pull changes from a remote repository.
func (s *Server) handlePull(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	slog.InfoContext(ctx, "Pulling changes", "id", id)

	repo, err := s.getRepoByID(id)
	if err != nil {
		slog.WarnContext(ctx, "Pull failed - repository not found", "id", id)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.Pull(repo.Path); err != nil {
		slog.ErrorContext(ctx, "Pull failed", "id", id, "path", repo.Path, "error", err)
		http.Error(w, "Failed to pull: "+err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Changes pulled successfully", "id", id, "path", repo.Path)
	w.WriteHeader(http.StatusOK)
}

// getRepoByID is a helper function to find a repository by its ID.
func (s *Server) getRepoByID(id string) (*models.Repository, error) {
	repos, err := s.store.LoadRepositories()
	if err != nil {
		return nil, err
	}
	for _, r := range repos {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, http.ErrMissingFile // Just a sentinel error
}
