package api

import (
	"encoding/json"
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

func (s *Server) handleAddUser(w http.ResponseWriter, r *http.Request) {
	var req AddUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	users, err := s.store.LoadUsers()
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}

	for _, user := range users {
		if user.Username == req.Username {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
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
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}

	// Don't return the password hash
	newUser.PasswordHash = ""
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

func (s *Server) handleRemoveUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	users, err := s.store.LoadUsers()
	if err != nil {
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
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := s.store.SaveUsers(newUsers); err != nil {
		http.Error(w, "Failed to save users", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.LoadUsers()
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}

	// Sanitize output
	safeUsers := []models.User{}
	for _, u := range users {
		u.PasswordHash = ""
		safeUsers = append(safeUsers, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeUsers)
}
