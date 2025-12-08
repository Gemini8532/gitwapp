package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gemini8532/gitwapp/frontend"
	"github.com/Gemini8532/gitwapp/internal/api"
	"github.com/Gemini8532/gitwapp/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const AppName = "GitWapp"

func main() {
	// Initialize Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load Configuration
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// Database Connection
	// For skeleton/dev purposes, fallback to SQLite if MySQL config is missing or connection fails
	// In production as per design, this should be MySQL.
	var db *gorm.DB
	var err error

	dsn := os.Getenv("MYSQL_DSN") // e.g., user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	if dsn != "" {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			logger.Error("Failed to connect to MySQL, falling back to SQLite", "error", err)
		}
	}

	if db == nil {
		// Ensure data directory exists
		home, _ := os.UserHomeDir()
		dataDir := filepath.Join(home, ".config", strings.ToLower(AppName))
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			logger.Error("Failed to create data dir", "error", err)
			os.Exit(1)
		}
		dbPath := filepath.Join(dataDir, "data.db")

		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			logger.Error("Failed to connect to SQLite", "error", err)
			os.Exit(1)
		}
		logger.Info("Connected to SQLite", "path", dbPath)
	} else {
		logger.Info("Connected to MySQL")
	}

	// Auto Migrate
	if err := models.Migrate(db); err != nil {
		logger.Error("Failed to migrate database", "error", err)
		os.Exit(1)
	}

	// Setup Router
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// API Routes
	handler := api.NewHandler(db)
	api.RegisterRoutes(r, handler)

	// Serve Frontend
	// We need to strip the "dist" prefix from the embedded fs
	distFS, err := fs.Sub(frontend.Dist, "dist")
	if err != nil {
		logger.Error("Failed to sub embedded fs", "error", err)
		os.Exit(1)
	}

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// If it's an API route that wasn't matched, return 404 JSON
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
			return
		}

		// Check if file exists in distFS
		_, err := distFS.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			c.FileFromFS(path, http.FS(distFS))
			return
		}

		// Otherwise serve index.html for SPA routing
		c.FileFromFS("/", http.FS(distFS))
	})

	logger.Info("Starting server", "port", port)
	if err := r.Run(":" + port); err != nil {
		logger.Error("Server failed", "error", err)
	}
}
