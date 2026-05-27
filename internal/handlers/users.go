package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
)

// user represents a user in the system
type user struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// userStore manages in-memory user storage
type userStore struct {
	mu    sync.RWMutex
	users map[string]user
}

// newUserStore creates a new user store with sample data
func newUserStore() *userStore {
	store := &userStore{
		users: make(map[string]user),
	}
	// add some sample users
	store.users["1"] = user{ID: "1", Name: "Alice", Email: "alice@example.com"}
	store.users["2"] = user{ID: "2", Name: "Bob", Email: "bob@example.com"}
	store.users["3"] = user{ID: "3", Name: "Charlie", Email: "charlie@example.com"}
	return store
}

// getAll returns all users as a slice
func (s *userStore) getAll() []user {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]user, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}

// global user store instance
var defaultUserStore = newUserStore()

// GetUsers returns all users as json
func GetUsers(w http.ResponseWriter, r *http.Request) {
	users := defaultUserStore.getAll()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
