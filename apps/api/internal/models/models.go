package models

import (
	"time"

	"github.com/google/uuid"
)

// Note represents a captured text note
type Note struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags"`
	Origin      Origin    `json:"origin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	SyncVersion int64     `json:"sync_version"`
	IsDeleted   bool      `json:"is_deleted"`
}

// Origin represents where the text was captured from
type Origin struct {
	Type      string `json:"type"` // "url", "book", "manual", "unknown"
	URL       string `json:"url,omitempty"`
	Title     string `json:"title,omitempty"`
	BookTitle string `json:"book_title,omitempty"`
	Chapter   string `json:"chapter,omitempty"`
	Page      string `json:"page,omitempty"`
	RawInput  string `json:"raw_input,omitempty"`
}

// Tag represents a note tag
type Tag struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// User represents an authenticated user
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SyncRequest represents a client sync request
type SyncRequest struct {
	LastSyncVersion int64  `json:"last_sync_version"`
	Notes           []Note `json:"notes"`
	Tags            []Tag  `json:"tags"`
}

// SyncResponse represents the server's sync response
type SyncResponse struct {
	ServerVersion int64  `json:"server_version"`
	Notes         []Note `json:"notes"`
	Tags          []Tag  `json:"tags"`
	Conflicts     []Note `json:"conflicts,omitempty"`
}
