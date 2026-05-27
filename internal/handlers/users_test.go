package handlers

import (
	"bytes"
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
	res := httptest.NewRecorder()

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

func TestCreateUserReturns201(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"David","email":"david@example.com"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	if res.Code != http.StatusCreated {
		t.Errorf("want status 201, got %d", res.Code)
	}
}

func TestCreateUserReturnsJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"David","email":"david@example.com"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	contentType := res.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("want 'application/json', got '%s'", contentType)
	}
}

func TestCreateUserAddsToStorage(t *testing.T) {
	originalStore := defaultUserStore
	defaultUserStore = newUserStore()
	defer func() { defaultUserStore = originalStore }()

	body := bytes.NewBufferString(`{"name":"Eve","email":"eve@example.com"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	var created user
	err := json.NewDecoder(res.Body).Decode(&created)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// verify user exists in store
	allUsers := defaultUserStore.getAll()
	found := false
	for _, u := range allUsers {
		if u.ID == created.ID && u.Name == "Eve" && u.Email == "eve@example.com" {
			found = true
			break
		}
	}

	if !found {
		t.Error("created user not found in storage")
	}
}

func TestCreateUserReturnsCreatedUser(t *testing.T) {
	originalStore := defaultUserStore
	defaultUserStore = newUserStore()
	defer func() { defaultUserStore = originalStore }()

	body := bytes.NewBufferString(`{"name":"Frank","email":"frank@example.com"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	var created user
	err := json.NewDecoder(res.Body).Decode(&created)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if created.Name != "Frank" {
		t.Errorf("want name 'Frank', got '%s'", created.Name)
	}
	if created.Email != "frank@example.com" {
		t.Errorf("want email 'frank@example.com', got '%s'", created.Email)
	}
	if created.ID == "" {
		t.Error("expected user id to be set")
	}
}

func TestCreateUserValidatesRequiredName(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"test@example.com"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("want status 400, got %d", res.Code)
	}

	var response map[string]string
	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["error"] != "name is required" {
		t.Errorf("want 'name is required', got '%s'", response["error"])
	}
}

func TestCreateUserValidatesRequiredEmail(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"Grace"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("want status 400, got %d", res.Code)
	}

	var response map[string]string
	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["error"] != "email is required" {
		t.Errorf("want 'email is required', got '%s'", response["error"])
	}
}

func TestCreateUserValidatesEmptyBody(t *testing.T) {
	body := bytes.NewBufferString(``)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("want status 400, got %d", res.Code)
	}
}

func TestCreateUserValidatesInvalidJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"Henry","email":}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("want status 400, got %d", res.Code)
	}
}
