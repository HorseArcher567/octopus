package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/models"
	"github.com/HorseArcher567/octopus/pkg/database"
)

// OrderRepository 订单数据访问接口
type OrderRepository interface {
	GetByID(ctx context.Context, orderID int64) (*models.Order, error)
	Create(ctx context.Context, order *models.Order) (int64, error)
}

type orderRepository struct {
	db *database.DB
}

// NewOrderRepository 创建订单仓库实例
func NewOrderRepository(db *database.DB) OrderRepository {
	return &orderRepository{db: db}
}

// GetByID 根据ID获取订单
func (r *orderRepository) GetByID(ctx context.Context, orderID int64) (*models.Order, error) {
	var order models.Order
	query := `SELECT order_id, user_id, product_name, amount, status, created_at, updated_at FROM orders WHERE order_id = ?`
	err := r.db.GetContext(ctx, &order, query, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}

// Create 创建订单
func (r *orderRepository) Create(ctx context.Context, order *models.Order) (int64, error) {
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
