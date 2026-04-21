package order

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) GetByID(ctx context.Context, id int64) (*Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, userID int64, productName string, amount float64) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("user_id must be positive: %w", ErrInvalidArgument)
	}
	if strings.TrimSpace(productName) == "" {
		return 0, fmt.Errorf("product_name is required: %w", ErrInvalidArgument)
	}
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be positive: %w", ErrInvalidArgument)
	}
	return s.repo.Create(ctx, &Order{UserID: userID, ProductName: productName, Amount: amount, Status: "pending"})
}
