package domain

import "time"

type Order struct {
	OrderID     int64
	UserID      int64
	ProductName string
	Amount      float64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
