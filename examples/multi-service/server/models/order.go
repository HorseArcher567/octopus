package models

import "time"

// Order 订单数据模型
type Order struct {
	OrderID     int64     `db:"order_id"`
	UserID      int64     `db:"user_id"`
	ProductName string    `db:"product_name"`
	Amount      float64   `db:"amount"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
