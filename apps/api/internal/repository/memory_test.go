package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/glean/api/internal/models"
)

func TestMemoryRepository_Users(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Test CreateUser
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hash123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Test GetUserByEmail
	found, err := repo.GetUserByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if found.ID != user.ID {
		t.Errorf("Expected user ID %v, got %v", user.ID, found.ID)
	}

	// Test GetUserByID
	found, err = repo.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if found.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, found.Email)
	}

	// Test not found
	_, err = repo.GetUserByEmail(ctx, "nonexistent@example.com")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMemoryRepository_Notes(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	userID := uuid.New()

	// Test CreateNote
	note := &models.Note{
		ID:          uuid.New(),
		UserID:      userID,
		Content:     "Test note content",
		Tags:        []string{"test", "example"},
		Origin:      models.Origin{Type: "url", URL: "https://example.com"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		SyncVersion: 1,
		IsDeleted:   false,
	}

	err := repo.CreateNote(ctx, note)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	// Test GetNotes
	notes, err := repo.GetNotes(ctx, userID)
	if err != nil {
		t.Fatalf("GetNotes failed: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}
	if notes[0].Content != "Test note content" {
		t.Errorf("Expected content 'Test note content', got '%s'", notes[0].Content)
	}

	// Test GetNote
	found, err := repo.GetNote(ctx, userID, note.ID)
	if err != nil {
		t.Fatalf("GetNote failed: %v", err)
	}
	if found.ID != note.ID {
		t.Errorf("Expected note ID %v, got %v", note.ID, found.ID)
	}

	// Test UpdateNote
	note.Content = "Updated content"
	note.SyncVersion = 2
	err = repo.UpdateNote(ctx, note)
	if err != nil {
		t.Fatalf("UpdateNote failed: %v", err)
	}

	found, _ = repo.GetNote(ctx, userID, note.ID)
	if found.Content != "Updated content" {
		t.Errorf("Expected updated content, got '%s'", found.Content)
	}

	// Test DeleteNote
	err = repo.DeleteNote(ctx, userID, note.ID)
	if err != nil {
		t.Fatalf("DeleteNote failed: %v", err)
	}

	notes, _ = repo.GetNotes(ctx, userID)
	if len(notes) != 0 {
		t.Errorf("Expected 0 notes after delete, got %d", len(notes))
	}
}

func TestMemoryRepository_Tags(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	userID := uuid.New()

	// Test CreateTag
	tag := &models.Tag{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      "work",
		Color:     "#ff0000",
		CreatedAt: time.Now(),
	}

	err := repo.CreateTag(ctx, tag)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	// Test GetTags
	tags, err := repo.GetTags(ctx, userID)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("Expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "work" {
		t.Errorf("Expected tag name 'work', got '%s'", tags[0].Name)
	}

	// Test DeleteTag
	err = repo.DeleteTag(ctx, userID, tag.ID)
	if err != nil {
		t.Fatalf("DeleteTag failed: %v", err)
	}

	tags, _ = repo.GetTags(ctx, userID)
	if len(tags) != 0 {
		t.Errorf("Expected 0 tags after delete, got %d", len(tags))
	}
}

func TestMemoryRepository_GetNotesSince(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	userID := uuid.New()

	// Create notes with different sync versions
	for i := 1; i <= 5; i++ {
		note := &models.Note{
			ID:          uuid.New(),
			UserID:      userID,
			Content:     "Note content",
			Tags:        []string{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			SyncVersion: int64(i),
			IsDeleted:   false,
		}
		repo.CreateNote(ctx, note)
	}

	// Get notes since version 3
	notes, err := repo.GetNotesSince(ctx, userID, 3)
	if err != nil {
		t.Fatalf("GetNotesSince failed: %v", err)
	}
	if len(notes) != 2 {
		t.Errorf("Expected 2 notes with version > 3, got %d", len(notes))
	}

	for _, note := range notes {
		if note.SyncVersion <= 3 {
			t.Errorf("Expected sync version > 3, got %d", note.SyncVersion)
		}
	}
}
