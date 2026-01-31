package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/glean/api/internal/models"
)

// Repository defines the interface for data persistence
type Repository interface {
	// Users
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// Notes
	GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error)
	GetNote(ctx context.Context, userID, noteID uuid.UUID) (*models.Note, error)
	CreateNote(ctx context.Context, note *models.Note) error
	UpdateNote(ctx context.Context, note *models.Note) error
	DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error
	GetNotesSince(ctx context.Context, userID uuid.UUID, syncVersion int64) ([]models.Note, error)

	// Tags
	GetTags(ctx context.Context, userID uuid.UUID) ([]models.Tag, error)
	CreateTag(ctx context.Context, tag *models.Tag) error
	DeleteTag(ctx context.Context, userID, tagID uuid.UUID) error

	// Health
	Ping(ctx context.Context) error
	Close() error
}
