package migrations

import (
	"gofr.dev/pkg/gofr/migration"
)

func setupInitialSchemas() migration.Migrate {
	return migration.Migrate{
		UP: func(d migration.Datasource) error {
			if _, err := d.SQL.Exec(`CREATE TABLE securities (
										id INT PRIMARY KEY AUTO_INCREMENT,
										isin VARCHAR(50) NOT NULL,
										symbol VARCHAR(50) NOT NULL,
										exchange INT NOT NULL,
										industry INT NOT NULL,
										name VARCHAR(100) NOT NULL,
										image VARCHAR(200) NOT NULL,
										ltp DECIMAL(10,2) NOT NULL,
										created_at TIMESTAMP NOT NULL,
										updated_at TIMESTAMP NOT NULL
									);`); err != nil {
				return err
			}

			if _, err := d.SQL.Exec(`CREATE UNIQUE INDEX idx_securities_symbol_exchange ON securities(symbol, exchange);`); err != nil {
				return err
			}

			if _, err := d.SQL.Exec(`CREATE TABLE universes (
										id INT PRIMARY KEY AUTO_INCREMENT,
										user_id INT,
										name VARCHAR(100) NOT NULL,
										created_at TIMESTAMP NOT NULL,
										updated_at TIMESTAMP NOT NULL
									);`); err != nil {
				return err
			}

			if _, err := d.SQL.Exec(`CREATE INDEX idx_universes_user_id ON universes(user_id);`); err != nil {
				return err
			}

			if _, err := d.SQL.Exec(`CREATE TABLE universe_security_mapping (
    									id INT PRIMARY KEY AUTO_INCREMENT,
										universe_id INT NOT NULL,
										security_id INT NOT NULL 
									);`); err != nil {
				return err
			}

			return nil
		},
	}
}
