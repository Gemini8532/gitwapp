package api

import (
	"net/http"
	"os"
	"time"

	"github.com/Gemini8532/gitwapp/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

// Auth Types
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Repo Types
type CreateRepoRequest struct {
	Path string `json:"path" binding:"required"`
}

// Handlers

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock auth for skeleton - accept any user/password, or better, check DB
	// For skeleton, we'll just check if user exists, if not create one?
	// Or just hardcode admin/admin for now to make it easy to test

	if req.Username == "admin" && req.Password == "admin" {
		// Create token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": req.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

		// Use a better secret in prod
		secret := []byte(os.Getenv("JWT_SECRET"))
		if len(secret) == 0 {
			secret = []byte("secret")
		}

		tokenString, err := token.SignedString(secret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}

		c.JSON(http.StatusOK, LoginResponse{Token: tokenString})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
}

func (h *Handler) GetRepos(c *gin.Context) {
	var repos []models.Repository
	if result := h.DB.Find(&repos); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, repos)
}

func (h *Handler) CreateRepo(c *gin.Context) {
	var req CreateRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real app, we would validate the path exists and is a git repo here
	// and get the directory name as the name.

	repo := models.Repository{
		Name:      req.Path, // Simplification
		Path:      req.Path,
		CreatedAt: time.Now(),
		UserID:    1, // Mock user ID
	}

	if result := h.DB.Create(&repo); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, repo)
}

func (h *Handler) DeleteRepo(c *gin.Context) {
	id := c.Param("id")
	if result := h.DB.Delete(&models.Repository{}, id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetRepoStatus(c *gin.Context) {
	id := c.Param("id")
	var repo models.Repository
	if result := h.DB.First(&repo, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Basic check using go-git
	r, err := git.PlainOpen(repo.Path)
	if err != nil {
		// If we can't open it, maybe it's not a git repo or path is invalid
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
		return
	}

	w, err := r.Worktree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	status, err := w.Status()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	isClean := status.IsClean()
	statusStr := "clean"
	if !isClean {
		statusStr = "dirty"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": statusStr,
		"clean": isClean,
		"file_count": len(status),
	})
}
