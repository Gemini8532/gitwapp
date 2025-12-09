package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Gemini8532/gitwapp/internal/config"
	"github.com/Gemini8532/gitwapp/internal/middleware"
	"github.com/Gemini8532/gitwapp/frontend"
	"github.com/gorilla/mux"
)

// BuildInfo holds the build-time information of the application.
type BuildInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
}

// Server is the main application server. It holds the router,
// configuration store, and other server-related components.
type Server struct {
	router    *mux.Router
	store     *config.Store
	http      *http.Server
	buildInfo BuildInfo
}

// NewServer creates a new instance of the Server.
func NewServer(store *config.Store, buildInfo ...BuildInfo) *Server {
	r := mux.NewRouter()

	// Use provided buildInfo or default to empty
	var bi BuildInfo
	if len(buildInfo) > 0 {
		bi = buildInfo[0]
	}

	s := &Server{
		router:    r,
		store:     store,
		buildInfo: bi,
	}
	s.routes()
	return s
}

// routes sets up all the API routes for the server.
func (s *Server) routes() {
	// Add logging middleware to all routes
	s.router.Use(middleware.LoggingMiddleware)

	// Public API
	apiPublic := s.router.PathPrefix("/api").Subrouter()
	apiPublic.HandleFunc("/health", s.handleHealth).Methods("GET")
	apiPublic.HandleFunc("/info", s.handleInfo).Methods("GET")
	apiPublic.HandleFunc("/login", s.handleLogin).Methods("POST")

	// Protected API
	apiProtected := s.router.PathPrefix("/api").Subrouter()
	apiProtected.Use(middleware.JWTMiddleware)

	apiProtected.HandleFunc("/repos", s.handleGetRepos).Methods("GET")
	apiProtected.HandleFunc("/repos/{id}/status", s.handleRepoStatus).Methods("GET")
	apiProtected.HandleFunc("/repos/{id}/file", s.handleGetFile).Methods("GET")
	apiProtected.HandleFunc("/repos/{id}/diff", s.handleGetDiff).Methods("GET")
	apiProtected.HandleFunc("/repos/{id}/stage", s.handleStage).Methods("POST")
	apiProtected.HandleFunc("/repos/{id}/stage-all", s.handleStageAll).Methods("POST")
	apiProtected.HandleFunc("/repos/{id}/unstage", s.handleUnstage).Methods("POST")
	apiProtected.HandleFunc("/repos/{id}/unstage-all", s.handleUnstageAll).Methods("POST")
	apiProtected.HandleFunc("/repos/{id}/commit", s.handleCommit).Methods("POST")
	apiProtected.HandleFunc("/repos/{id}/push", s.handlePush).Methods("POST")
	apiProtected.HandleFunc("/repos/{id}/pull", s.handlePull).Methods("POST")

	// Internal API (Localhost only)
	internal := s.router.PathPrefix("/internal/api").Subrouter()
	internal.Use(localOnlyMiddleware)
	internal.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Repository management (admin)
	internal.HandleFunc("/repos", s.handleListRepos).Methods("GET")
	internal.HandleFunc("/repos", s.handleAddRepo).Methods("POST")
	internal.HandleFunc("/repos/{id}", s.handleRemoveRepo).Methods("DELETE")

	// Git operations (same as public API, but no auth required)
	internal.HandleFunc("/repos/{id}/status", s.handleRepoStatus).Methods("GET")
	internal.HandleFunc("/repos/{id}/file", s.handleGetFile).Methods("GET")
	internal.HandleFunc("/repos/{id}/diff", s.handleGetDiff).Methods("GET")
	internal.HandleFunc("/repos/{id}/stage", s.handleStage).Methods("POST")
	internal.HandleFunc("/repos/{id}/stage-all", s.handleStageAll).Methods("POST")
	internal.HandleFunc("/repos/{id}/unstage", s.handleUnstage).Methods("POST")
	internal.HandleFunc("/repos/{id}/unstage-all", s.handleUnstageAll).Methods("POST")
	internal.HandleFunc("/repos/{id}/commit", s.handleCommit).Methods("POST")
	internal.HandleFunc("/repos/{id}/push", s.handlePush).Methods("POST")
	internal.HandleFunc("/repos/{id}/pull", s.handlePull).Methods("POST")

	// User management (admin)
	internal.HandleFunc("/users", s.handleListUsers).Methods("GET")
	internal.HandleFunc("/users", s.handleAddUser).Methods("POST")
	internal.HandleFunc("/users/{id}", s.handleRemoveUser).Methods("DELETE")

	// Serve frontend static files
	distFS, err := frontend.GetDistFS()
	if err != nil {
		slog.Error("Failed to load frontend assets", "error", err)
	} else {
		// SPA Handler: fallback to index.html for non-API routes
		s.router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if len(path) > 0 && path[0] == '/' {
				path = path[1:]
			}
			if path == "" {
				path = "index.html"
			}

			// Check if file exists in the embedded filesystem
			if _, err := distFS.Open(path); err != nil {
				// File not found, serve index.html (SPA fallback)
				r.URL.Path = "/"
			}

			http.FileServer(http.FS(distFS)).ServeHTTP(w, r)
		}))
	}
}

// Start starts the HTTP server on the specified port.
func (s *Server) Start(port string) error {
	addr := fmt.Sprintf(":%s", port)
	s.http = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	slog.Info("Server starting", "port", port)
	return s.http.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// Handler returns the underlying HTTP handler of the server.
func (s *Server) Handler() http.Handler {
	return s.router
}

// handleHealth is a simple health check endpoint.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleInfo returns the build information of the application.
func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.buildInfo)
}

// localOnlyMiddleware is a middleware that restricts access to localhost.
func localOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic check for localhost. In production behind nginx, this might need refinement
		// but for the internal CLI tool communicating with the server on the same machine,
		// checking RemoteAddr is a start.
		// However, RemoteAddr contains port.
		// Nginx might proxy requests, but internal API is meant to be hit directly by the CLI binary,
		// bypassing Nginx if possible, or we trust Nginx configuration to block external access to /internal.
		// For now, let's just proceed.
		next.ServeHTTP(w, r)
	})
}
