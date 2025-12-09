package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Gemini8532/gitwapp/pkg/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type AddUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (s *Server) handleAddUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AddUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "Failed to decode add user request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "Adding user", "username", req.Username, "password_length", len(req.Password))

	if req.Username == "" || req.Password == "" {
		slog.WarnContext(ctx, "Add user failed - username and password are required")
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	users, err := s.store.LoadUsers()
	if err != nil {
		slog.ErrorContext(ctx, "Add user failed - unable to load existing users", "username", req.Username, "error", err)
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}

	for _, user := range users {
		if user.Username == req.Username {
			slog.WarnContext(ctx, "Add user failed - username already exists", "username", req.Username)
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.ErrorContext(ctx, "Add user failed - unable to hash password", "username", req.Username, "error", err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	newUser := models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	users = append(users, newUser)

	if err := s.store.SaveUsers(users); err != nil {
		slog.ErrorContext(ctx, "Add user failed - unable to save user", "username", req.Username, "id", newUser.ID, "error", err)
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "User saved with password hash", "id", newUser.ID, "username", req.Username, "hash_length", len(newUser.PasswordHash))

	// Return response without password hash
	response := UserResponse{
		ID:       newUser.ID,
		Username: newUser.Username,
	}
	slog.InfoContext(ctx, "User added successfully", "id", newUser.ID, "username", req.Username)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleRemoveUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	slog.InfoContext(ctx, "Removing user", "id", id)

	users, err := s.store.LoadUsers()
	if err != nil {
		slog.ErrorContext(ctx, "Remove user failed - unable to load users", "id", id, "error", err)
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}

	newUsers := []models.User{}
	found := false
	for _, user := range users {
		if user.ID == id {
			found = true
			continue
		}
		newUsers = append(newUsers, user)
	}

	if !found {
		slog.WarnContext(ctx, "Remove user failed - user not found", "id", id)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := s.store.SaveUsers(newUsers); err != nil {
		slog.ErrorContext(ctx, "Remove user failed - unable to save users", "id", id, "error", err)
		http.Error(w, "Failed to save users", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "User removed successfully", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slog.InfoContext(ctx, "Listing users")
	users, err := s.store.LoadUsers()
	if err != nil {
		slog.ErrorContext(ctx, "List users failed - unable to load users", "error", err)
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}

	// Sanitize output
	safeUsers := []models.User{}
	for _, u := range users {
		u.PasswordHash = ""
		safeUsers = append(safeUsers, u)
	}

	slog.InfoContext(ctx, "Users listed successfully", "count", len(safeUsers))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeUsers)
}
