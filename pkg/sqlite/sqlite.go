package sqlite

import (
	"fmt"

	"github.com/HorseArcher567/octopus/pkg/database"
	_ "modernc.org/sqlite"
)

func New(cfg *Config) (*database.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("sqlite: config cannot be nil")
	}
	cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return database.Open("sqlite", cfg.DSN, cfg.Pool)
}

func MustNew(cfg *Config) *database.DB {
	db, err := New(cfg)
	if err != nil {
		panic("sqlite: failed to create DB: " + err.Error())
	}
	return db
}
