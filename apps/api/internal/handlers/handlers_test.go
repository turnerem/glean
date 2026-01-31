package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glean/api/internal/middleware"
	"github.com/glean/api/internal/models"
	"github.com/glean/api/internal/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() *gin.Engine {
	// Use in-memory repository for tests
	repo := repository.NewMemoryRepository()
	SetRepository(repo)

	r := gin.New()

	// Auth routes
	r.POST("/auth/register", Register)
	r.POST("/auth/login", Login)

	return r
}

func setupFullTestRouter() *gin.Engine {
	// Use in-memory repository for tests
	repo := repository.NewMemoryRepository()
	SetRepository(repo)

	r := gin.New()

	// Auth routes (public)
	r.POST("/auth/register", Register)
	r.POST("/auth/login", Login)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.Auth())
	{
		protected.GET("/notes", GetNotes)
		protected.POST("/notes", CreateNote)
		protected.PUT("/notes/:id", UpdateNote)
		protected.DELETE("/notes/:id", DeleteNote)
		protected.GET("/tags", GetTags)
		protected.POST("/tags", CreateTag)
		protected.POST("/sync", Sync)
	}

	return r
}

func registerAndGetToken(t *testing.T, router *gin.Engine, email string) string {
	body := map[string]string{
		"email":    email,
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to register user: %s", w.Body.String())
	}

	var response AuthResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	return response.Token
}

func TestRegister(t *testing.T) {
	router := setupTestRouter()

	// Test successful registration
	body := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token in response")
	}
	if response.User == nil {
		t.Error("Expected user in response")
	}
}

func TestRegisterValidation(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name     string
		body     map[string]string
		expected int
	}{
		{
			name:     "missing email",
			body:     map[string]string{"password": "password123"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "missing password",
			body:     map[string]string{"email": "test@example.com"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "invalid email",
			body:     map[string]string{"email": "notanemail", "password": "password123"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "short password",
			body:     map[string]string{"email": "test@example.com", "password": "short"},
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d: %s", tt.expected, w.Code, w.Body.String())
			}
		})
	}
}

func TestLogin(t *testing.T) {
	router := setupTestRouter()

	// First register a user
	registerBody := map[string]string{
		"email":    "login@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to register user: %s", w.Body.String())
	}

	// Test successful login
	loginBody := map[string]string{
		"email":    "login@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token in response")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	router := setupTestRouter()

	// Register a user first
	registerBody := map[string]string{
		"email":    "invalid@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Test wrong password
	loginBody := map[string]string{
		"email":    "invalid@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ = json.Marshal(loginBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// Test non-existent user
	loginBody = map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	router := setupTestRouter()

	body := map[string]string{
		"email":    "duplicate@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	// First registration
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("First registration failed: %s", w.Body.String())
	}

	// Second registration with same email
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 for duplicate email, got %d", w.Code)
	}
}

func TestCreateNote(t *testing.T) {
	router := setupFullTestRouter()
	token := registerAndGetToken(t, router, "notes@example.com")

	body := map[string]interface{}{
		"content": "Test note content",
		"tags":    []string{"tag1", "tag2"},
		"origin": map[string]string{
			"type": "url",
			"url":  "https://example.com",
		},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var note models.Note
	if err := json.Unmarshal(w.Body.Bytes(), &note); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if note.Content != "Test note content" {
		t.Errorf("Expected content 'Test note content', got '%s'", note.Content)
	}
	if len(note.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(note.Tags))
	}
}

func TestGetNotes(t *testing.T) {
	router := setupFullTestRouter()
	token := registerAndGetToken(t, router, "getnotes@example.com")

	// Create a note first
	createBody := map[string]interface{}{
		"content": "Note to retrieve",
		"tags":    []string{},
	}
	jsonBody, _ := json.Marshal(createBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Get notes
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/notes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var notes []models.Note
	if err := json.Unmarshal(w.Body.Bytes(), &notes); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("Expected 1 note, got %d", len(notes))
	}
}

func TestUpdateNote(t *testing.T) {
	router := setupFullTestRouter()
	token := registerAndGetToken(t, router, "update@example.com")

	// Create a note
	createBody := map[string]interface{}{
		"content": "Original content",
		"tags":    []string{},
	}
	jsonBody, _ := json.Marshal(createBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	var createdNote models.Note
	json.Unmarshal(w.Body.Bytes(), &createdNote)

	// Update the note
	updateBody := map[string]interface{}{
		"content": "Updated content",
		"tags":    []string{"updated"},
	}
	jsonBody, _ = json.Marshal(updateBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/notes/"+createdNote.ID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var updatedNote models.Note
	json.Unmarshal(w.Body.Bytes(), &updatedNote)

	if updatedNote.Content != "Updated content" {
		t.Errorf("Expected 'Updated content', got '%s'", updatedNote.Content)
	}
}

func TestDeleteNote(t *testing.T) {
	router := setupFullTestRouter()
	token := registerAndGetToken(t, router, "delete@example.com")

	// Create a note
	createBody := map[string]interface{}{
		"content": "Note to delete",
		"tags":    []string{},
	}
	jsonBody, _ := json.Marshal(createBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	var createdNote models.Note
	json.Unmarshal(w.Body.Bytes(), &createdNote)

	// Delete the note
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/notes/"+createdNote.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify note is deleted
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/notes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	var notes []models.Note
	json.Unmarshal(w.Body.Bytes(), &notes)

	if len(notes) != 0 {
		t.Errorf("Expected 0 notes after delete, got %d", len(notes))
	}
}

func TestCreateTag(t *testing.T) {
	router := setupFullTestRouter()
	token := registerAndGetToken(t, router, "tags@example.com")

	body := map[string]string{
		"name":  "work",
		"color": "#ff0000",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/tags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var tag models.Tag
	if err := json.Unmarshal(w.Body.Bytes(), &tag); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if tag.Name != "work" {
		t.Errorf("Expected name 'work', got '%s'", tag.Name)
	}
}

func TestGetTags(t *testing.T) {
	router := setupFullTestRouter()
	token := registerAndGetToken(t, router, "gettags@example.com")

	// Create a tag first
	createBody := map[string]string{
		"name":  "personal",
		"color": "#00ff00",
	}
	jsonBody, _ := json.Marshal(createBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/tags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Get tags
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/tags", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var tags []models.Tag
	if err := json.Unmarshal(w.Body.Bytes(), &tags); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}
}

func TestSync(t *testing.T) {
	router := setupFullTestRouter()
	token := registerAndGetToken(t, router, "sync@example.com")

	// Create a sync request with empty data
	body := models.SyncRequest{
		LastSyncVersion: 0,
		Notes:           []models.Note{},
		Tags:            []models.Tag{},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/sync", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response models.SyncResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	router := setupFullTestRouter()

	// Try to access protected route without token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/notes", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestUpdateNonExistentNote(t *testing.T) {
	router := setupFullTestRouter()
	token := registerAndGetToken(t, router, "nonexistent@example.com")

	body := map[string]string{
		"content": "Updated content",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/notes/00000000-0000-0000-0000-000000000000", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInvalidNoteID(t *testing.T) {
	router := setupFullTestRouter()
	token := registerAndGetToken(t, router, "invalidid@example.com")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/notes/not-a-uuid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
