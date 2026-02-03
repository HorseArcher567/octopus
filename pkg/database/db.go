package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// DB wraps a sqlx.DB instance with unified configuration and initialization.
// It embeds *sqlx.DB so all sqlx methods are directly available.
type DB struct {
	*sqlx.DB
}

// New creates a new database connection with the given configuration.
func New(cfg *Config) (*DB, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Open database connection using sqlx
	db, err := sqlx.Open(cfg.GetDriverName(), cfg.DSN)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	db.SetConnMaxLifetime(cfg.GetConnMaxLifetime())
	db.SetConnMaxIdleTime(cfg.GetConnMaxIdleTime())

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return &DB{DB: db}, nil
}

// MustNew creates a new database connection and panics if an error occurs.
func MustNew(cfg *Config) *DB {
	db, err := New(cfg)
	if err != nil {
		panic("database: failed to create DB: " + err.Error())
	}
	return db
}

// Ping verifies the connection to the database is still alive with a timeout.
func (db *DB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

// Stats returns database connection pool statistics.
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}
