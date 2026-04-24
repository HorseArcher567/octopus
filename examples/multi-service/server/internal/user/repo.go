package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/HorseArcher567/octopus/pkg/database"
)

type Repository interface {
	GetByID(ctx context.Context, userID int64) (*User, error)
	Create(ctx context.Context, user *User) (int64, error)
}

type repository struct {
	db *database.DB
}

type userRecord struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewRepository(db *database.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, userID int64) (*User, error) {
	var rec userRecord
	query := `SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?`
	err := r.db.GetContext(ctx, &rec, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user %d: %w", userID, ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &User{ID: rec.ID, Username: rec.Username, Email: rec.Email, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt}, nil
}

func (r *repository) Create(ctx context.Context, user *User) (int64, error) {
	query := `INSERT INTO users (username, email, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	result, err := r.db.ExecContext(ctx, query, user.Username, user.Email)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}
	return id, nil
}
