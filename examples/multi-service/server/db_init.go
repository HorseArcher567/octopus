package main

import (
	"context"
	"fmt"

	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

// initDatabase 初始化数据库并创建表结构
func initDatabase(db *database.DB) error {
	// 创建表结构
	if err := createTables(context.Background(), db); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// createTables 创建数据库表
func createTables(ctx context.Context, db *database.DB) error {
	tables := []string{
		// 用户表
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_username (username),
			INDEX idx_email (email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		// 产品表
		`CREATE TABLE IF NOT EXISTS products (
			product_id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL,
			stock INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		// 订单表
		`CREATE TABLE IF NOT EXISTS orders (
			order_id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			product_name VARCHAR(255) NOT NULL,
			amount DECIMAL(10, 2) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_user_id (user_id),
			INDEX idx_status (status),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}

	for _, tableSQL := range tables {
		if _, err := db.ExecContext(ctx, tableSQL); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	log := xlog.FromContext(ctx)
	log.Info("database tables created successfully")
	return nil
}
