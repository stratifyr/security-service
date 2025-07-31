package migrations

import (
	"gofr.dev/pkg/gofr/migration"
)

func All() map[int64]migration.Migrate {
	return map[int64]migration.Migrate{
		1742025361: setupInitialSchemas(),
		1753808918: addTierField(),
	}
}
