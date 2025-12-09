// Package config provides a simple file-based storage solution for the
// GitWapp application's configuration, including users and repositories.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Gemini8532/gitwapp/pkg/models"
)

const (
	ConfigDirName = "gitwapp"
	UsersFile     = "users.json"
	ReposFile     = "repositories.json"
	PIDFile       = "gitwapp.pid"
)

// Store provides a thread-safe way to manage application configuration
// stored in JSON files.
type Store struct {
	configDir string
	mu        sync.RWMutex
}

// NewStore creates a new Store and initializes the configuration directory
// in the default user config location.
func NewStore() (*Store, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config dir: %w", err)
	}

	configDir := filepath.Join(userConfigDir, ConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	return &Store{
		configDir: configDir,
	}, nil
}

// NewStoreWithDir creates a new Store with a custom directory path.
// This is useful for testing or running in a sandboxed environment.
func NewStoreWithDir(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}
	return &Store{
		configDir: dir,
	}, nil
}

// GetUsersPath returns the full path to the users JSON file.
func (s *Store) GetUsersPath() string {
	return filepath.Join(s.configDir, UsersFile)
}

// GetReposPath returns the full path to the repositories JSON file.
func (s *Store) GetReposPath() string {
	return filepath.Join(s.configDir, ReposFile)
}

// GetPIDFilePath returns the full path to the PID file.
func (s *Store) GetPIDFilePath() string {
	return filepath.Join(s.configDir, PIDFile)
}

// LoadUsers reads the users from the users.json file.
func (s *Store) LoadUsers() ([]models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.GetUsersPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []models.User{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var users []models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// SaveUsers writes the users to the users.json file.
func (s *Store) SaveUsers(users []models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.GetUsersPath(), data, 0644)
}

// LoadRepositories reads the repositories from the repositories.json file.
func (s *Store) LoadRepositories() ([]models.Repository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.GetReposPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []models.Repository{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var repos []models.Repository
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// SaveRepositories writes the repositories to the repositories.json file.
func (s *Store) SaveRepositories(repos []models.Repository) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.GetReposPath(), data, 0644)
}
