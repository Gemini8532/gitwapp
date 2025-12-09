package api

import (
	"encoding/json"
	"net/http"

	"github.com/Gemini8532/gitwapp/internal/git"
	"github.com/Gemini8532/gitwapp/pkg/models"
	"github.com/gorilla/mux"
)

// Public API: Get all repos
func (s *Server) handleGetRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := s.store.LoadRepositories()
	if err != nil {
		http.Error(w, "Failed to load repositories", http.StatusInternalServerError)
		return
	}

	// TODO: Maybe enrich with simple status (clean/dirty) if performance allows
	// For now, just return the list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repos)
}

// Public API: Get repo detailed status
func (s *Server) handleRepoStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	repo, err := s.getRepoByID(id)
	if err != nil {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	status, err := git.GetStatus(repo.Path)
	if err != nil {
		http.Error(w, "Failed to get git status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Public API: Stage file
type StageRequest struct {
	File string `json:"file"`
}

func (s *Server) handleStage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req StageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	repo, err := s.getRepoByID(id)
	if err != nil {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.StageFile(repo.Path, req.File); err != nil {
		http.Error(w, "Failed to stage file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Public API: Commit
type CommitRequest struct {
	Message string `json:"message"`
}

func (s *Server) handleCommit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req CommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Commit message required", http.StatusBadRequest)
		return
	}

	repo, err := s.getRepoByID(id)
	if err != nil {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.Commit(repo.Path, req.Message); err != nil {
		http.Error(w, "Failed to commit: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Public API: Push
func (s *Server) handlePush(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	repo, err := s.getRepoByID(id)
	if err != nil {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.Push(repo.Path); err != nil {
		http.Error(w, "Failed to push: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Public API: Pull
func (s *Server) handlePull(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	repo, err := s.getRepoByID(id)
	if err != nil {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := git.Pull(repo.Path); err != nil {
		http.Error(w, "Failed to pull: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Helper
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
