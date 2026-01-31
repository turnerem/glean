package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/glean/api/internal/middleware"
	"github.com/glean/api/internal/models"
	"github.com/glean/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var repo repository.Repository

// SetRepository sets the repository for handlers to use
func SetRepository(r repository.Repository) {
	repo = r
}

// Auth handlers

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Check if user exists
	_, err := repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}
	if !errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := repo.CreateUser(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate token
	token, err := middleware.GenerateToken(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{Token: token, User: user})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Find user
	user, err := repo.GetUserByEmail(ctx, req.Email)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate token
	token, err := middleware.GenerateToken(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{Token: token, User: user})
}

// Note handlers

type CreateNoteRequest struct {
	Content string        `json:"content" binding:"required"`
	Tags    []string      `json:"tags"`
	Origin  models.Origin `json:"origin"`
}

type UpdateNoteRequest struct {
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func getUserID(c *gin.Context) uuid.UUID {
	userIDStr := c.GetString("user_id")
	userID, _ := uuid.Parse(userIDStr)
	return userID
}

func GetNotes(c *gin.Context) {
	userID := getUserID(c)
	ctx := c.Request.Context()

	notes, err := repo.GetNotes(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notes"})
		return
	}

	c.JSON(http.StatusOK, notes)
}

func CreateNote(c *gin.Context) {
	userID := getUserID(c)
	ctx := c.Request.Context()

	var req CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note := &models.Note{
		ID:          uuid.New(),
		UserID:      userID,
		Content:     req.Content,
		Tags:        req.Tags,
		Origin:      req.Origin,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		SyncVersion: 1,
		IsDeleted:   false,
	}

	if note.Tags == nil {
		note.Tags = []string{}
	}

	if err := repo.CreateNote(ctx, note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

func UpdateNote(c *gin.Context) {
	userID := getUserID(c)
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	ctx := c.Request.Context()

	var req UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing note
	note, err := repo.GetNote(ctx, userID, noteID)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Update fields
	if req.Content != "" {
		note.Content = req.Content
	}
	if req.Tags != nil {
		note.Tags = req.Tags
	}
	note.UpdatedAt = time.Now()
	note.SyncVersion++

	if err := repo.UpdateNote(ctx, note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

func DeleteNote(c *gin.Context) {
	userID := getUserID(c)
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	ctx := c.Request.Context()

	err = repo.DeleteNote(ctx, userID, noteID)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted"})
}

// Tag handlers

type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

func GetTags(c *gin.Context) {
	userID := getUserID(c)
	ctx := c.Request.Context()

	tags, err := repo.GetTags(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tags"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

func CreateTag(c *gin.Context) {
	userID := getUserID(c)
	ctx := c.Request.Context()

	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag := &models.Tag{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      req.Name,
		Color:     req.Color,
		CreatedAt: time.Now(),
	}

	if err := repo.CreateTag(ctx, tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// Sync handler

func Sync(c *gin.Context) {
	userID := getUserID(c)
	ctx := c.Request.Context()

	var req models.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get server notes since client's last sync
	serverNotes, err := repo.GetNotesSince(ctx, userID, req.LastSyncVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notes"})
		return
	}

	// Process client notes
	conflicts := []models.Note{}
	for _, clientNote := range req.Notes {
		clientNote.UserID = userID

		existingNote, err := repo.GetNote(ctx, userID, clientNote.ID)
		if errors.Is(err, repository.ErrNotFound) {
			// New note from client
			if err := repo.CreateNote(ctx, &clientNote); err != nil {
				continue
			}
		} else if err == nil {
			// Existing note - check for conflicts
			if clientNote.SyncVersion > existingNote.SyncVersion {
				// Client has newer version
				clientNote.SyncVersion = existingNote.SyncVersion + 1
				repo.UpdateNote(ctx, &clientNote)
			} else if clientNote.SyncVersion < existingNote.SyncVersion {
				// Server has newer - mark as conflict
				conflicts = append(conflicts, *existingNote)
			}
		}
	}

	// Process client tags
	for _, clientTag := range req.Tags {
		clientTag.UserID = userID
		repo.CreateTag(ctx, &clientTag) // Ignore errors for duplicate tags
	}

	// Get all server data to return
	allNotes, _ := repo.GetNotes(ctx, userID)
	allTags, _ := repo.GetTags(ctx, userID)

	// Calculate max sync version
	var maxVersion int64
	for _, n := range allNotes {
		if n.SyncVersion > maxVersion {
			maxVersion = n.SyncVersion
		}
	}

	c.JSON(http.StatusOK, models.SyncResponse{
		ServerVersion: maxVersion,
		Notes:         serverNotes,
		Tags:          allTags,
		Conflicts:     conflicts,
	})
}

// Health check with context
func Health(ctx context.Context) error {
	return repo.Ping(ctx)
}
