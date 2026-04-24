package user

import (
	"context"
	"database/sql"
	"errors"
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

func TestRepositoryGetByIDNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRepository(db)
	query := `SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(1001)).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByID(context.Background(), 1001)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestRepositoryCreateSuccess(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRepository(db)
	query := `INSERT INTO users (username, email, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs("alice", "alice@example.com").WillReturnResult(sqlmock.NewResult(88, 1))
	id, err := repo.Create(context.Background(), &User{Username: "alice", Email: "alice@example.com"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id != 88 {
		t.Fatalf("expected id=88, got %d", id)
	}
}

func TestListShapeCompileGuard(_ *testing.T) {
	_ = time.Date
}
