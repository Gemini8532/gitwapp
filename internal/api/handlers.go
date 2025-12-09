package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/Gemini8532/gitwapp/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Store *Storage
}

func NewHandler(store *Storage) *Handler {
	return &Handler{Store: store}
}

// LoginHandler handles user authentication
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.Store.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(config.JWTSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{Token: tokenString})
}

// ListReposHandler returns all tracked repositories
func (h *Handler) ListReposHandler(w http.ResponseWriter, r *http.Request) {
	repos, err := h.Store.LoadRepos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(repos)
}

// GetRepoStatusHandler returns the detailed git status of a repository
func (h *Handler) GetRepoStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	repos, err := h.Store.LoadRepos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var repo *Repository
	for _, r := range repos {
		if r.ID == id {
			repo = &r
			break
		}
	}

	if repo == nil {
		http.Error(w, "Repo not found", http.StatusNotFound)
		return
	}

	rRepo, err := git.PlainOpen(repo.Path)
	if err != nil {
		http.Error(w, "Failed to open git repo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	wTree, err := rRepo.Worktree()
	if err != nil {
		http.Error(w, "Failed to get worktree: "+err.Error(), http.StatusInternalServerError)
		return
	}

	status, err := wTree.Status()
	if err != nil {
		http.Error(w, "Failed to get status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	headRef, err := rRepo.Head()
	if err != nil {
		http.Error(w, "Failed to get HEAD: " + err.Error(), http.StatusInternalServerError)
		return
	}

	branchName := headRef.Name().Short()

	// Simplified status calculation
	isClean := status.IsClean()

	// Ahead/Behind logic (very basic for now)
	ahead, behind := 0, 0

    // Attempt to calculate ahead/behind if remote exists
    // This is a simplified check and might need more robust logic
    remoteName := "origin"
    remotes, err := rRepo.Remotes()
    if err == nil && len(remotes) > 0 {
         // Check if origin exists
         hasOrigin := false
         for _, remote := range remotes {
             if remote.Config().Name == remoteName {
                 hasOrigin = true
                 break
             }
         }

         if hasOrigin {
             // We need to resolve local branch hash and remote branch hash
             // Assuming tracking is set up correctly, or guessing refs/remotes/origin/<branchName>

             localHash := headRef.Hash()

             remoteRefName := plumbing.ReferenceName("refs/remotes/" + remoteName + "/" + branchName)
             remoteRef, err := rRepo.Reference(remoteRefName, true)
             if err == nil {
                 remoteHash := remoteRef.Hash()

                 // Calculate commits ahead/behind
                 // This requires traversing the commit history which can be expensive
                 // Using a simplified approach: just checking if hashes are different
                 if localHash != remoteHash {
                     // TODO: Implement proper traversal to count ahead/behind
                     // For MVP, if different, we mark as potentially diverged or just indicate mismatch
                     // To properly count, we need to iterate log

                     // Just a placeholder for now as full graph traversal is complex to implement in one go
                     // without helpers
                 }
             }
         }
    }


	json.NewEncoder(w).Encode(RepoStatus{
		IsClean:       isClean,
		HasChanges:    !isClean,
		Ahead:         ahead,
		Behind:        behind,
		CurrentBranch: branchName,
	})
}


// Internal API Handlers

func (h *Handler) InternalAddRepoHandler(w http.ResponseWriter, r *http.Request) {
    // Expects JSON body with path
    var req struct {
        Path string `json:"path"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    absPath, err := filepath.Abs(req.Path)
    if err != nil {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }

    // Check if it's a valid git repo
    _, err = git.PlainOpen(absPath)
    if err != nil {
         http.Error(w, "Not a valid git repository: " + err.Error(), http.StatusBadRequest)
         return
    }

    repoName := filepath.Base(absPath)

    newRepo := Repository{
        ID: uuid.New().String(),
        Name: repoName,
        Path: absPath,
        CreatedAt: time.Now(),
        UserID: "system", // default for now
    }

    if err := h.Store.AddRepo(newRepo); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(newRepo)
}

func (h *Handler) InternalAddUserHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
         http.Error(w, "Failed to hash password", http.StatusInternalServerError)
         return
    }

    newUser := User{
        ID: uuid.New().String(),
        Username: req.Username,
        PasswordHash: string(hashedPassword),
    }

    if err := h.Store.AddUser(newUser); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(newUser)
}
