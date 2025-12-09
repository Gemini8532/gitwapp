package api

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Storage handles persistence of users and repositories
type Storage struct {
	UserFile string
	RepoFile string
	mu       sync.Mutex
}

func NewStorage(userFile, repoFile string) *Storage {
	return &Storage{
		UserFile: userFile,
		RepoFile: repoFile,
	}
}

func (s *Storage) LoadUsers() ([]User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.UserFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []User{}, nil
		}
		return nil, err
	}

	var users []User
	if len(data) == 0 {
		return []User{}, nil
	}
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Storage) SaveUsers(users []User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.UserFile, data, 0644)
}

func (s *Storage) LoadRepos() ([]Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.RepoFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Repository{}, nil
		}
		return nil, err
	}

	var repos []Repository
	if len(data) == 0 {
		return []Repository{}, nil
	}
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func (s *Storage) SaveRepos(repos []Repository) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.RepoFile, data, 0644)
}

func (s *Storage) AddUser(u User) error {
	users, err := s.LoadUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Username == u.Username {
			return fmt.Errorf("user already exists")
		}
	}
	users = append(users, u)
	return s.SaveUsers(users)
}

func (s *Storage) GetUserByUsername(username string) (*User, error) {
	users, err := s.LoadUsers()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.Username == username {
			return &user, nil
		}
	}
	return nil, nil
}

func (s *Storage) AddRepo(r Repository) error {
	repos, err := s.LoadRepos()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		if repo.Path == r.Path {
			return fmt.Errorf("repo path already tracked")
		}
	}
	repos = append(repos, r)
	return s.SaveRepos(repos)
}

func (s *Storage) RemoveRepo(id string) error {
	repos, err := s.LoadRepos()
	if err != nil {
		return err
	}
	newRepos := []Repository{}
	found := false
	for _, repo := range repos {
		if repo.ID == id {
			found = true
			continue
		}
		newRepos = append(newRepos, repo)
	}
	if !found {
		return fmt.Errorf("repo not found")
	}
	return s.SaveRepos(newRepos)
}
