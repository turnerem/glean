package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/glean/api/internal/models"
)

// MemoryRepository is an in-memory implementation for development/testing
type MemoryRepository struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*models.User
	notes map[uuid.UUID][]models.Note
	tags  map[uuid.UUID][]models.Tag
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users: make(map[uuid.UUID]*models.User),
		notes: make(map[uuid.UUID][]models.Note),
		tags:  make(map[uuid.UUID][]models.Tag),
	}
}

func (r *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}

func (r *MemoryRepository) Close() error {
	return nil
}

// Users

func (r *MemoryRepository) CreateUser(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for existing email
	for _, u := range r.users {
		if u.Email == user.Email {
			return ErrNotFound // Use a better error in production
		}
	}

	r.users[user.ID] = user
	return nil
}

func (r *MemoryRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrNotFound
}

func (r *MemoryRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return user, nil
}

// Notes

func (r *MemoryRepository) GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userNotes := r.notes[userID]
	result := make([]models.Note, 0)
	for _, n := range userNotes {
		if !n.IsDeleted {
			result = append(result, n)
		}
	}
	return result, nil
}

func (r *MemoryRepository) GetNote(ctx context.Context, userID, noteID uuid.UUID) (*models.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userNotes := r.notes[userID]
	for _, n := range userNotes {
		if n.ID == noteID {
			return &n, nil
		}
	}
	return nil, ErrNotFound
}

func (r *MemoryRepository) CreateNote(ctx context.Context, note *models.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.notes[note.UserID] = append(r.notes[note.UserID], *note)
	return nil
}

func (r *MemoryRepository) UpdateNote(ctx context.Context, note *models.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userNotes := r.notes[note.UserID]
	for i, n := range userNotes {
		if n.ID == note.ID {
			userNotes[i] = *note
			r.notes[note.UserID] = userNotes
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userNotes := r.notes[userID]
	for i, n := range userNotes {
		if n.ID == noteID {
			userNotes[i].IsDeleted = true
			userNotes[i].SyncVersion++
			r.notes[userID] = userNotes
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) GetNotesSince(ctx context.Context, userID uuid.UUID, syncVersion int64) ([]models.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userNotes := r.notes[userID]
	result := make([]models.Note, 0)
	for _, n := range userNotes {
		if n.SyncVersion > syncVersion {
			result = append(result, n)
		}
	}
	return result, nil
}

// Tags

func (r *MemoryRepository) GetTags(ctx context.Context, userID uuid.UUID) ([]models.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userTags := r.tags[userID]
	if userTags == nil {
		return []models.Tag{}, nil
	}
	return userTags, nil
}

func (r *MemoryRepository) CreateTag(ctx context.Context, tag *models.Tag) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tags[tag.UserID] = append(r.tags[tag.UserID], *tag)
	return nil
}

func (r *MemoryRepository) DeleteTag(ctx context.Context, userID, tagID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userTags := r.tags[userID]
	for i, t := range userTags {
		if t.ID == tagID {
			r.tags[userID] = append(userTags[:i], userTags[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}
