package user

import (
	"context"
	"fmt"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id int64) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, username, email string) (int64, error) {
	if strings.TrimSpace(username) == "" {
		return 0, fmt.Errorf("username is required: %w", ErrInvalidArgument)
	}
	if strings.TrimSpace(email) == "" {
		return 0, fmt.Errorf("email is required: %w", ErrInvalidArgument)
	}
	return s.repo.Create(ctx, &User{Username: username, Email: email})
}
