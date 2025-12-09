package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Gemini8532/gitwapp/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "Failed to decode login request", "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "Login attempt", "username", req.Username)

	users, err := s.store.LoadUsers()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to load users for login", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, user := range users {
		if user.Username == req.Username {
			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err == nil {
				// Success
				token, err := middleware.GenerateToken(user.ID, user.Username)
				if err != nil {
					slog.ErrorContext(ctx, "Failed to generate JWT token", "error", err, "user_id", user.ID)
					http.Error(w, "Failed to generate token", http.StatusInternalServerError)
					return
				}

				// Set cookie as well for easier frontend handling if needed,
				// but returning JSON is primary.
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    token,
					HttpOnly: true,
					Path:     "/",
					Secure:   false, // Set to true in prod with HTTPS
				})

				slog.InfoContext(ctx, "User logged in successfully", "user_id", user.ID, "username", user.Username)

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(LoginResponse{Token: token})
				return
			} else {
				slog.InfoContext(ctx, "Invalid password provided", "username", req.Username)
			}
		}
	}

	slog.InfoContext(ctx, "Login failed - invalid credentials", "username", req.Username)
	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}
