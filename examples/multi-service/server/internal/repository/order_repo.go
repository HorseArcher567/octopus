package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/domain"
	"github.com/HorseArcher567/octopus/pkg/database"
)

type OrderRepository interface {
	GetByID(ctx context.Context, orderID int64) (*domain.Order, error)
	Create(ctx context.Context, order *domain.Order) (int64, error)
}

type orderRepository struct {
	db *database.DB
}

type orderRecord struct {
	OrderID     int64     `db:"order_id"`
	UserID      int64     `db:"user_id"`
	ProductName string    `db:"product_name"`
	Amount      float64   `db:"amount"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func NewOrderRepository(db *database.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetByID(ctx context.Context, orderID int64) (*domain.Order, error) {
	var rec orderRecord
	query := `SELECT order_id, user_id, product_name, amount, status, created_at, updated_at FROM orders WHERE order_id = ?`
	err := r.db.GetContext(ctx, &rec, query, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("order %d: %w", orderID, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return toOrderDomain(rec), nil
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) (int64, error) {
	query := `INSERT INTO orders (user_id, product_name, amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`
	result, err := r.db.ExecContext(ctx, query, order.UserID, order.ProductName, order.Amount, order.Status)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}
	return id, nil
}

func toOrderDomain(rec orderRecord) *domain.Order {
	return &domain.Order{
		OrderID:     rec.OrderID,
		UserID:      rec.UserID,
		ProductName: rec.ProductName,
		Amount:      rec.Amount,
		Status:      rec.Status,
		CreatedAt:   rec.CreatedAt,
		UpdatedAt:   rec.UpdatedAt,
	}
}
