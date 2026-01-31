package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/glean/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(databaseURL string) (*PostgresRepository, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *PostgresRepository) Close() error {
	r.pool.Close()
	return nil
}

// Users

func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &user, err
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &user, err
}

// Notes

func (r *PostgresRepository) GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, content, tags, origin, created_at, updated_at, sync_version, is_deleted
		 FROM notes WHERE user_id = $1 AND is_deleted = false
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNotes(rows)
}

func (r *PostgresRepository) GetNote(ctx context.Context, userID, noteID uuid.UUID) (*models.Note, error) {
	var note models.Note
	var tagsJSON, originJSON []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, content, tags, origin, created_at, updated_at, sync_version, is_deleted
		 FROM notes WHERE id = $1 AND user_id = $2`,
		noteID, userID,
	).Scan(
		&note.ID, &note.UserID, &note.Content, &tagsJSON, &originJSON,
		&note.CreatedAt, &note.UpdatedAt, &note.SyncVersion, &note.IsDeleted,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(tagsJSON, &note.Tags)
	json.Unmarshal(originJSON, &note.Origin)

	return &note, nil
}

func (r *PostgresRepository) CreateNote(ctx context.Context, note *models.Note) error {
	tagsJSON, _ := json.Marshal(note.Tags)
	originJSON, _ := json.Marshal(note.Origin)

	_, err := r.pool.Exec(ctx,
		`INSERT INTO notes (id, user_id, content, tags, origin, created_at, updated_at, sync_version, is_deleted)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		note.ID, note.UserID, note.Content, tagsJSON, originJSON,
		note.CreatedAt, note.UpdatedAt, note.SyncVersion, note.IsDeleted,
	)
	return err
}

func (r *PostgresRepository) UpdateNote(ctx context.Context, note *models.Note) error {
	tagsJSON, _ := json.Marshal(note.Tags)

	result, err := r.pool.Exec(ctx,
		`UPDATE notes SET content = $1, tags = $2, updated_at = $3, sync_version = $4
		 WHERE id = $5 AND user_id = $6`,
		note.Content, tagsJSON, note.UpdatedAt, note.SyncVersion, note.ID, note.UserID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE notes SET is_deleted = true, updated_at = NOW(), sync_version = sync_version + 1
		 WHERE id = $1 AND user_id = $2`,
		noteID, userID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) GetNotesSince(ctx context.Context, userID uuid.UUID, syncVersion int64) ([]models.Note, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, content, tags, origin, created_at, updated_at, sync_version, is_deleted
		 FROM notes WHERE user_id = $1 AND sync_version > $2
		 ORDER BY sync_version ASC`,
		userID, syncVersion,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNotes(rows)
}

// Tags

func (r *PostgresRepository) GetTags(ctx context.Context, userID uuid.UUID) ([]models.Tag, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, color, created_at
		 FROM tags WHERE user_id = $1
		 ORDER BY name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if tags == nil {
		tags = []models.Tag{}
	}
	return tags, rows.Err()
}

func (r *PostgresRepository) CreateTag(ctx context.Context, tag *models.Tag) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tags (id, user_id, name, color, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		tag.ID, tag.UserID, tag.Name, tag.Color, tag.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) DeleteTag(ctx context.Context, userID, tagID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM tags WHERE id = $1 AND user_id = $2`,
		tagID, userID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Helper functions

func scanNotes(rows pgx.Rows) ([]models.Note, error) {
	var notes []models.Note
	for rows.Next() {
		var note models.Note
		var tagsJSON, originJSON []byte

		err := rows.Scan(
			&note.ID, &note.UserID, &note.Content, &tagsJSON, &originJSON,
			&note.CreatedAt, &note.UpdatedAt, &note.SyncVersion, &note.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(tagsJSON, &note.Tags)
		json.Unmarshal(originJSON, &note.Origin)

		if note.Tags == nil {
			note.Tags = []string{}
		}

		notes = append(notes, note)
	}

	if notes == nil {
		notes = []models.Note{}
	}
	return notes, rows.Err()
}
