package order

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

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

func TestRepositoryGetByIDNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRepository(db)
	query := `SELECT order_id, user_id, product_name, amount, status, created_at, updated_at FROM orders WHERE order_id = ?`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(2001)).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByID(context.Background(), 2001)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}
