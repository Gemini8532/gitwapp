package api

import (
	"encoding/json"
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
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	users, err := s.store.LoadUsers()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, user := range users {
		if user.Username == req.Username {
			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err == nil {
				// Success
				token, err := middleware.GenerateToken(user.ID, user.Username)
				if err != nil {
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

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(LoginResponse{Token: token})
				return
			}
		}
	}

	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}
