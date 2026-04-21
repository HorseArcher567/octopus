package product

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
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

func TestRepositoryListDefaultPaging(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRepository(db)
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
}
