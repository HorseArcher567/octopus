package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/domain"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/jmoiron/sqlx"
)

func newMockDB(t *testing.T) (*database.DB, sqlmock.Sqlmock) {
	t.Helper()
	rawDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = rawDB.Close() })
	sqlxDB := sqlx.NewDb(rawDB, "sqlmock")
	return &database.DB{DB: sqlxDB}, mock
}

func TestUserRepositoryGetByIDNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)

	query := `SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(1001)).WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 1001)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepositoryCreateSuccess(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)

	query := `INSERT INTO users (username, email, created_at, updated_at) VALUES (?, ?, NOW(), NOW())`
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs("alice", "alice@example.com").WillReturnResult(sqlmock.NewResult(88, 1))

	id, err := repo.Create(context.Background(), &domain.User{Username: "alice", Email: "alice@example.com"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id != 88 {
		t.Fatalf("expected id=88, got %d", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestOrderRepositoryGetByIDNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)

	query := `SELECT order_id, user_id, product_name, amount, status, created_at, updated_at FROM orders WHERE order_id = ?`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(2001)).WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 2001)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestProductRepositoryListDefaultPaging(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewProductRepository(db)

	countQuery := `SELECT COUNT(*) FROM products`
	mock.ExpectQuery(regexp.QuoteMeta(countQuery)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	listQuery := `SELECT product_id, name, description, price, stock, created_at, updated_at FROM products ORDER BY product_id LIMIT ? OFFSET ?`
	rows := sqlmock.NewRows([]string{"product_id", "name", "description", "price", "stock", "created_at", "updated_at"})
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rows.AddRow(int64(1), "book", "good", 9.9, 10, now, now)
	rows.AddRow(int64(2), "pen", "blue", 1.2, 20, now, now)
	mock.ExpectQuery(regexp.QuoteMeta(listQuery)).WithArgs(int32(10), int32(0)).WillReturnRows(rows)

	items, total, err := repo.List(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("unexpected list result total=%d len=%d", total, len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
