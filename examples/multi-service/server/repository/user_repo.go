package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/models"
	"github.com/HorseArcher567/octopus/pkg/database"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	GetByID(ctx context.Context, userID int64) (*models.User, error)
	Create(ctx context.Context, user *models.User) (int64, error)
}

type userRepository struct {
	db *database.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{db: db}
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?`
	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *models.User) (int64, error) {
	query := `INSERT INTO users (username, email, created_at, updated_at) VALUES (?, ?, NOW(), NOW())`
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
