package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/models"
	"github.com/HorseArcher567/octopus/pkg/database"
)

// ProductRepository 产品数据访问接口
type ProductRepository interface {
	GetByID(ctx context.Context, productID int64) (*models.Product, error)
	List(ctx context.Context, page, pageSize int32) ([]*models.Product, int64, error)
}

type productRepository struct {
	db *database.DB
}

// NewProductRepository 创建产品仓库实例
func NewProductRepository(db *database.DB) ProductRepository {
	return &productRepository{db: db}
}

// GetByID 根据ID获取产品
func (r *productRepository) GetByID(ctx context.Context, productID int64) (*models.Product, error) {
	var product models.Product
	query := `SELECT product_id, name, description, price, stock, created_at, updated_at FROM products WHERE product_id = ?`
	err := r.db.GetContext(ctx, &product, query, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}

// List 分页获取产品列表
func (r *productRepository) List(ctx context.Context, page, pageSize int32) ([]*models.Product, int64, error) {
	// 计算偏移量
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM products`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// 查询列表
	var products []*models.Product
	query := `SELECT product_id, name, description, price, stock, created_at, updated_at FROM products ORDER BY product_id LIMIT ? OFFSET ?`
	err = r.db.SelectContext(ctx, &products, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}
