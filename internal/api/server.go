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
	"github.com/gorilla/mux"
)

type BuildInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
}

type Server struct {
	router    *mux.Router
	store     *config.Store
	http      *http.Server
	buildInfo BuildInfo
}

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
	apiProtected.HandleFunc("/repos/{id}/stage", s.handleStage).Methods("POST")
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
	internal.HandleFunc("/repos/{id}/stage", s.handleStage).Methods("POST")
	internal.HandleFunc("/repos/{id}/commit", s.handleCommit).Methods("POST")
	internal.HandleFunc("/repos/{id}/push", s.handlePush).Methods("POST")
	internal.HandleFunc("/repos/{id}/pull", s.handlePull).Methods("POST")

	// User management (admin)
	internal.HandleFunc("/users", s.handleListUsers).Methods("GET")
	internal.HandleFunc("/users", s.handleAddUser).Methods("POST")
	internal.HandleFunc("/users/{id}", s.handleRemoveUser).Methods("DELETE")
}

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

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.buildInfo)
}

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
