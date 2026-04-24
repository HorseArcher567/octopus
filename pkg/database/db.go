package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const defaultPingTimeout = 5 * time.Second

// DB wraps a sqlx.DB instance with unified configuration and initialization.
// It embeds *sqlx.DB so all sqlx methods are directly available.
type DB struct {
	*sqlx.DB
}

// Open creates a new database connection using the provided driver, DSN, and pool settings.
func Open(driverName, dsn string, pool PoolConfig) (*DB, error) {
	if strings.TrimSpace(driverName) == "" {
		return nil, fmt.Errorf("database: driverName is required")
	}
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("database: DSN is required")
	}

	pool.Normalize()
	if err := pool.Validate(); err != nil {
		return nil, err
	}

	db, err := sqlx.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	applyPool(db, pool)
	wrapped := &DB{DB: db}
	if err := wrapped.PingTimeout(pool.PingTimeout); err != nil {
		_ = db.Close()
		return nil, err
	}
	return wrapped, nil
}

// MustOpen creates a new database connection and panics if an error occurs.
func MustOpen(driverName, dsn string, pool PoolConfig) *DB {
	db, err := Open(driverName, dsn, pool)
	if err != nil {
		panic("database: failed to open DB: " + err.Error())
	}
	return db
}

// Wrap applies pool configuration and ping checks to an existing sqlx.DB.
func Wrap(db *sqlx.DB, pool PoolConfig) (*DB, error) {
	if db == nil {
		return nil, fmt.Errorf("database: db cannot be nil")
	}
	pool.Normalize()
	if err := pool.Validate(); err != nil {
		return nil, err
	}
	applyPool(db, pool)
	wrapped := &DB{DB: db}
	if err := wrapped.PingTimeout(pool.PingTimeout); err != nil {
		return nil, err
	}
	return wrapped, nil
}

// MustWrap panics when Wrap returns an error.
func MustWrap(db *sqlx.DB, pool PoolConfig) *DB {
	wrapped, err := Wrap(db, pool)
	if err != nil {
		panic("database: failed to wrap DB: " + err.Error())
	}
	return wrapped
}

func applyPool(db *sqlx.DB, cfg PoolConfig) {
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
}

// Ping verifies the connection to the database is still alive with a timeout.
func (db *DB) Ping() error {
	return db.PingTimeout(0)
}

// PingTimeout verifies the connection to the database is still alive with a timeout.
func (db *DB) PingTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		timeout = defaultPingTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return db.PingContext(ctx)
}

// Stats returns database connection pool statistics.
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}
