package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/domain"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) Create(ctx context.Context, username, email string) (int64, error) {
	if strings.TrimSpace(username) == "" {
		return 0, fmt.Errorf("username is required: %w", domain.ErrInvalidArgument)
	}
	if strings.TrimSpace(email) == "" {
		return 0, fmt.Errorf("email is required: %w", domain.ErrInvalidArgument)
	}
	return s.repo.Create(ctx, &domain.User{Username: username, Email: email})
}

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *OrderService) Create(ctx context.Context, userID int64, productName string, amount float64) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("user_id must be positive: %w", domain.ErrInvalidArgument)
	}
	if strings.TrimSpace(productName) == "" {
		return 0, fmt.Errorf("product_name is required: %w", domain.ErrInvalidArgument)
	}
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be positive: %w", domain.ErrInvalidArgument)
	}
	return s.repo.Create(ctx, &domain.Order{UserID: userID, ProductName: productName, Amount: amount, Status: "pending"})
}

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) List(ctx context.Context, page, pageSize int32) ([]*domain.Product, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}
