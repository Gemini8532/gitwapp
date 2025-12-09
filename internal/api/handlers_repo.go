package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Gemini8532/gitwapp/pkg/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AddRepoRequest struct {
	Path string `json:"path"`
	Name string `json:"name"` // Optional, default to base of path
}

func (s *Server) handleListRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slog.InfoContext(ctx, "Listing repositories")
	repos, err := s.store.LoadRepositories()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to load repositories", "error", err)
		http.Error(w, "Failed to load repositories", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Repositories listed successfully", "count", len(repos))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repos)
}

func (s *Server) handleAddRepo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AddRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "Failed to decode add repo request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "Adding repository", "path", req.Path)

	if req.Path == "" {
		slog.WarnContext(ctx, "Add repository failed - path is required")
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	// Validate path exists and is a directory
	info, err := os.Stat(req.Path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.WarnContext(ctx, "Add repository failed - path does not exist", "path", req.Path)
			http.Error(w, "Path does not exist", http.StatusBadRequest)
			return
		}
		slog.ErrorContext(ctx, "Add repository failed - error accessing path", "path", req.Path, "error", err)
		http.Error(w, "Error accessing path", http.StatusInternalServerError)
		return
	}
	if !info.IsDir() {
		slog.WarnContext(ctx, "Add repository failed - path is not a directory", "path", req.Path)
		http.Error(w, "Path is not a directory", http.StatusBadRequest)
		return
	}

	// Load existing repos to check for duplicates
	repos, err := s.store.LoadRepositories()
	if err != nil {
		slog.ErrorContext(ctx, "Add repository failed - unable to load existing repositories", "error", err)
		http.Error(w, "Failed to load repositories", http.StatusInternalServerError)
		return
	}

	for _, repo := range repos {
		if repo.Path == req.Path {
			slog.WarnContext(ctx, "Add repository failed - repository already tracked", "path", req.Path)
			http.Error(w, "Repository already tracked", http.StatusConflict)
			return
		}
	}

	// Create new repo object
	newRepo := models.Repository{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Path:      req.Path,
		CreatedAt: time.Now(),
		// UserID: ... (To be implemented with Auth)
	}

	if newRepo.Name == "" {
		newRepo.Name = req.Path // specific logic to extract base name can be added later
	}

	repos = append(repos, newRepo)

	if err := s.store.SaveRepositories(repos); err != nil {
		slog.ErrorContext(ctx, "Add repository failed - unable to save repository", "path", req.Path, "error", err)
		http.Error(w, "Failed to save repository", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Repository added successfully", "id", newRepo.ID, "path", newRepo.Path)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newRepo)
}

func (s *Server) handleRemoveRepo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	slog.InfoContext(ctx, "Removing repository", "id", id)

	repos, err := s.store.LoadRepositories()
	if err != nil {
		slog.ErrorContext(ctx, "Remove repository failed - unable to load repositories", "id", id, "error", err)
		http.Error(w, "Failed to load repositories", http.StatusInternalServerError)
		return
	}

	newRepos := []models.Repository{}
	found := false
	for _, repo := range repos {
		if repo.ID == id {
			found = true
			continue
		}
		newRepos = append(newRepos, repo)
	}

	if !found {
		slog.WarnContext(ctx, "Remove repository failed - repository not found", "id", id)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := s.store.SaveRepositories(newRepos); err != nil {
		slog.ErrorContext(ctx, "Remove repository failed - unable to save repositories", "id", id, "error", err)
		http.Error(w, "Failed to save repositories", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Repository removed successfully", "id", id)
	w.WriteHeader(http.StatusNoContent)
}
