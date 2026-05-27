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

// createUserRequest represents the request body for creating a user
type createUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// userStore manages in-memory user storage
type userStore struct {
	mu     sync.RWMutex
	users  map[string]user
	nextID int
}

// newUserStore creates a new user store with sample data
func newUserStore() *userStore {
	store := &userStore{
		users:  make(map[string]user),
		nextID: 4,
	}
	//sample users will remove later
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

// create adds a new user to the store
func (s *userStore) create(name, email string) user {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID
	s.nextID++

	u := user{
		ID:    string(rune(id + 48)), // convert int to string
		Name:  name,
		Email: email,
	}
	s.users[u.ID] = u
	return u
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

// CreateUser creates a new user from json request body
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest

	// parse json request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid json body"})
		return
	}

	// validate required fields
	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name is required"})
		return
	}

	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "email is required"})
		return
	}

	// add user to storage
	newUser := defaultUserStore.create(req.Name, req.Email)

	// return created user with status 201
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}
