package mysql

import (
	"fmt"

	"github.com/HorseArcher567/octopus/pkg/database"
	_ "github.com/go-sql-driver/mysql"
)

func New(cfg *Config) (*database.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("mysql: config cannot be nil")
	}
	cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return database.Open("mysql", cfg.DSN, cfg.Pool)
}

func MustNew(cfg *Config) *database.DB {
	db, err := New(cfg)
	if err != nil {
		panic("mysql: failed to create DB: " + err.Error())
	}
	return db
}
