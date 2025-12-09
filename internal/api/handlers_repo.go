package api

import (
	"encoding/json"
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
	repos, err := s.store.LoadRepositories()
	if err != nil {
		http.Error(w, "Failed to load repositories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repos)
}

func (s *Server) handleAddRepo(w http.ResponseWriter, r *http.Request) {
	var req AddRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	// Validate path exists and is a directory
	info, err := os.Stat(req.Path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Path does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Error accessing path", http.StatusInternalServerError)
		return
	}
	if !info.IsDir() {
		http.Error(w, "Path is not a directory", http.StatusBadRequest)
		return
	}

	// Load existing repos to check for duplicates
	repos, err := s.store.LoadRepositories()
	if err != nil {
		http.Error(w, "Failed to load repositories", http.StatusInternalServerError)
		return
	}

	for _, repo := range repos {
		if repo.Path == req.Path {
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
		http.Error(w, "Failed to save repository", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newRepo)
}

func (s *Server) handleRemoveRepo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	repos, err := s.store.LoadRepositories()
	if err != nil {
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
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if err := s.store.SaveRepositories(newRepos); err != nil {
		http.Error(w, "Failed to save repositories", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
