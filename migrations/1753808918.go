package migrations

import "gofr.dev/pkg/gofr/migration"

func addTierField() migration.Migrate {
	return migration.Migrate{
		UP: func(d migration.Datasource) error {
			_, err := d.SQL.Exec(`ALTER TABLE securities ADD COLUMN tier INT NOT NULL DEFAULT 0 AFTER ltp;`)
			if err != nil {
				return err
			}

			_, err = d.SQL.Exec(`ALTER TABLE metrics ADD COLUMN tier INT NOT NULL DEFAULT 0 AFTER indicator;`)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
