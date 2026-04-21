package shared

import (
	"github.com/HorseArcher567/octopus/pkg/assemble"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/store"
)

const MySQLPrimary = "primary"

func PrimaryDB(ctx *assemble.Context) (*database.DB, error) {
	return store.GetNamed[*database.DB](ctx.Store(), MySQLPrimary)
}
