package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/HorseArcher567/octopus/pkg/database"
)

type Repository interface {
	GetByID(ctx context.Context, productID int64) (*Product, error)
	List(ctx context.Context, page, pageSize int32) ([]*Product, int64, error)
}

type repository struct{ db *database.DB }

type productRecord struct {
	ProductID   int64     `db:"product_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Price       float64   `db:"price"`
	Stock       int       `db:"stock"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func NewRepository(db *database.DB) Repository { return &repository{db: db} }

func (r *repository) GetByID(ctx context.Context, productID int64) (*Product, error) {
	var rec productRecord
	query := `SELECT product_id, name, description, price, stock, created_at, updated_at FROM products WHERE product_id = ?`
	err := r.db.GetContext(ctx, &rec, query, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, fmt.Errorf("product %d: %w", productID, ErrNotFound) }
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &Product{ProductID: rec.ProductID, Name: rec.Name, Description: rec.Description, Price: rec.Price, Stock: rec.Stock, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt}, nil
}

func (r *repository) List(ctx context.Context, page, pageSize int32) ([]*Product, int64, error) {
	offset := (page - 1) * pageSize
	if offset < 0 { offset = 0 }
	if pageSize <= 0 { pageSize = 10 }
	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM products`); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}
	var recs []productRecord
	query := `SELECT product_id, name, description, price, stock, created_at, updated_at FROM products ORDER BY product_id LIMIT ? OFFSET ?`
	if err := r.db.SelectContext(ctx, &recs, query, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	products := make([]*Product, 0, len(recs))
	for _, rec := range recs {
		products = append(products, &Product{ProductID: rec.ProductID, Name: rec.Name, Description: rec.Description, Price: rec.Price, Stock: rec.Stock, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt})
	}
	return products, total, nil
}
