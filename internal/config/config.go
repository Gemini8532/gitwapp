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

type Store struct {
	configDir string
	mu        sync.RWMutex
}

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

func NewStoreWithDir(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}
	return &Store{
		configDir: dir,
	}, nil
}

func (s *Store) GetUsersPath() string {
	return filepath.Join(s.configDir, UsersFile)
}

func (s *Store) GetReposPath() string {
	return filepath.Join(s.configDir, ReposFile)
}

func (s *Store) GetPIDFilePath() string {
	return filepath.Join(s.configDir, PIDFile)
}

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

func (s *Store) SaveUsers(users []models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.GetUsersPath(), data, 0644)
}

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

func (s *Store) SaveRepositories(repos []models.Repository) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.GetReposPath(), data, 0644)
}
