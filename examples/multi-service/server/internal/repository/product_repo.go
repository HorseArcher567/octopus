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

type ProductRepository interface {
	GetByID(ctx context.Context, productID int64) (*domain.Product, error)
	List(ctx context.Context, page, pageSize int32) ([]*domain.Product, int64, error)
}

type productRepository struct {
	db *database.DB
}

type productRecord struct {
	ProductID   int64     `db:"product_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Price       float64   `db:"price"`
	Stock       int       `db:"stock"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func NewProductRepository(db *database.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetByID(ctx context.Context, productID int64) (*domain.Product, error) {
	var rec productRecord
	query := `SELECT product_id, name, description, price, stock, created_at, updated_at FROM products WHERE product_id = ?`
	err := r.db.GetContext(ctx, &rec, query, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("product %d: %w", productID, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return toProductDomain(rec), nil
}

func (r *productRepository) List(ctx context.Context, page, pageSize int32) ([]*domain.Product, int64, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM products`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	var recs []productRecord
	query := `SELECT product_id, name, description, price, stock, created_at, updated_at FROM products ORDER BY product_id LIMIT ? OFFSET ?`
	err = r.db.SelectContext(ctx, &recs, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	products := make([]*domain.Product, 0, len(recs))
	for _, rec := range recs {
		products = append(products, toProductDomain(rec))
	}
	return products, total, nil
}

func toProductDomain(rec productRecord) *domain.Product {
	return &domain.Product{
		ProductID:   rec.ProductID,
		Name:        rec.Name,
		Description: rec.Description,
		Price:       rec.Price,
		Stock:       rec.Stock,
		CreatedAt:   rec.CreatedAt,
		UpdatedAt:   rec.UpdatedAt,
	}
}
