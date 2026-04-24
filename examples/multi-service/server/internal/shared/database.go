package shared

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/hook"
	"github.com/HorseArcher567/octopus/pkg/store"
)

const PrimaryDBName = "primary"

//go:embed embed/schema.sql
var schemaSQL string

func PrimaryDB(r store.Reader) (*database.DB, error) {
	return store.GetNamed[*database.DB](r, PrimaryDBName)
}

// InitSchema is an app-level startup hook that applies the example schema to
// the primary database. This is especially useful for in-memory SQLite, where
// schema must be created inside the running process instead of by an external
// CLI beforehand.
func InitSchema(h *hook.Context) error {
	db, err := PrimaryDB(h)
	if err != nil {
		return fmt.Errorf("get primary database: %w", err)
	}
	if strings.TrimSpace(schemaSQL) == "" {
		return fmt.Errorf("embedded schema.sql is empty")
	}
	if _, err := db.ExecContext(h.Context(), schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
