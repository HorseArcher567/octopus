package shared

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/assemble"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/store"
)

//go:embed embed/schema.sql
var schemaSQL string

// SetupSchema applies the example schema to the primary database.
// This is especially useful for in-memory SQLite, where schema must be created
// inside the running process instead of by an external CLI beforehand.
func SetupSchema() assemble.SetupStep {
	return assemble.SetupStep{
		Name: "schema",
		Run: func(ctx *assemble.SetupContext) error {
			db, err := store.GetNamed[*database.DB](ctx.Store(), PrimaryDBName)
			if err != nil {
				return fmt.Errorf("get primary database: %w", err)
			}
			if strings.TrimSpace(schemaSQL) == "" {
				return fmt.Errorf("embedded schema.sql is empty")
			}
			if _, err := db.Exec(schemaSQL); err != nil {
				return fmt.Errorf("apply schema: %w", err)
			}
			return nil
		},
	}
}
