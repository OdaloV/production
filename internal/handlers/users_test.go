package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUsersReturns200(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	GetUsers(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want status 200, got %d", res.Code)
	}
}

func TestGetUsersReturnsJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	res := httptstest.NewRecorder()

	GetUsers(res, req)

	contentType := res.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("want 'application/json', got '%s'", contentType)
	}
}

func TestGetUsersReturnsArray(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	GetUsers(res, req)

	var users []user
	err := json.NewDecoder(res.Body).Decode(&users)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if len(users) == 0 {
		t.Error("expected at least one user")
	}
}

func TestGetUsersReturnsCorrectStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	GetUsers(res, req)

	var users []user
	err := json.NewDecoder(res.Body).Decode(&users)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if len(users) < 3 {
		t.Errorf("expected at least 3 users, got %d", len(users))
	}

	for _, u := range users {
		if u.ID == "" {
			t.Error("user missing id field")
		}
		if u.Name == "" {
			t.Error("user missing name field")
		}
		if u.Email == "" {
			t.Error("user missing email field")
		}
	}
}

func TestGetUsersReturnsAllUsers(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	GetUsers(res, req)

	var users []user
	err := json.NewDecoder(res.Body).Decode(&users)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	// check that sample users exist
	foundIDs := make(map[string]bool)
	for _, u := range users {
		foundIDs[u.ID] = true
	}

	expectedIDs := []string{"1", "2", "3"}
	for _, id := range expectedIDs {
		if !foundIDs[id] {
			t.Errorf("expected user with id '%s' not found", id)
		}
	}
}

func TestGetUsersReturnsEmptyArrayWhenNoUsers(t *testing.T) {
	// save original store
	originalStore := defaultUserStore

	// create empty store
	emptyStore := &userStore{
		users: make(map[string]user),
	}
	defaultUserStore = emptyStore

	defer func() {
		defaultUserStore = originalStore
	}()

	req := httptest.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	GetUsers(res, req)

	var users []user
	err := json.NewDecoder(res.Body).Decode(&users)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("expected empty array, got %d users", len(users))
	}
}
