package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/HorseArcher567/octopus/pkg/database"
)

type Repository interface {
	GetByID(ctx context.Context, orderID int64) (*Order, error)
	Create(ctx context.Context, order *Order) (int64, error)
}

type repository struct{ db *database.DB }

type orderRecord struct {
	OrderID     int64     `db:"order_id"`
	UserID      int64     `db:"user_id"`
	ProductName string    `db:"product_name"`
	Amount      float64   `db:"amount"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func NewRepository(db *database.DB) Repository { return &repository{db: db} }

func (r *repository) GetByID(ctx context.Context, orderID int64) (*Order, error) {
	var rec orderRecord
	query := `SELECT order_id, user_id, product_name, amount, status, created_at, updated_at FROM orders WHERE order_id = ?`
	err := r.db.GetContext(ctx, &rec, query, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("order %d: %w", orderID, ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &Order{OrderID: rec.OrderID, UserID: rec.UserID, ProductName: rec.ProductName, Amount: rec.Amount, Status: rec.Status, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt}, nil
}

func (r *repository) Create(ctx context.Context, order *Order) (int64, error) {
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
