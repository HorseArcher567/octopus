package product

import "time"

type Product struct {
	ProductID   int64
	Name        string
	Description string
	Price       float64
	Stock       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
