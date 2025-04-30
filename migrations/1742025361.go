package migrations

import (
	"gofr.dev/pkg/gofr/migration"
)

func setupInitialSchemas() migration.Migrate {
	return migration.Migrate{
		UP: func(d migration.Datasource) error {
			if _, err := d.SQL.Exec(`CREATE TABLE securities (
										id INT PRIMARY KEY AUTO_INCREMENT,
										isin VARCHAR(50) NOT NULL UNIQUE,
										symbol VARCHAR(50) NOT NULL,
										industry INT NOT NULL,
										name VARCHAR(100) NOT NULL,
										image VARCHAR(200) NOT NULL,
										ltp DECIMAL(10,2) NOT NULL,
										created_at TIMESTAMP NOT NULL,
										updated_at TIMESTAMP NOT NULL
									);`); err != nil {
				return err
			}

			if _, err := d.SQL.Exec(`CREATE TABLE universes (
										id INT PRIMARY KEY AUTO_INCREMENT,
										user_id INT NOT NULL,
										name VARCHAR(100) NOT NULL,
										created_at TIMESTAMP NOT NULL,
										updated_at TIMESTAMP NOT NULL,
										
                                        INDEX idx_universes_user_id (user_id)
									);`); err != nil {
				return err
			}

			if _, err := d.SQL.Exec(`CREATE TABLE universe_securities (
                                        id INT PRIMARY KEY AUTO_INCREMENT,
										universe_id INT NOT NULL,
										security_id INT NOT NULL,
										status enum('ENABLED', 'DISABLED') NOT NULL,
										created_at TIMESTAMP NOT NULL,
										updated_at TIMESTAMP NOT NULL,
										
                                        CONSTRAINT fk_universe_securities_universe_id FOREIGN KEY (universe_id) REFERENCES universes(id),
                                        CONSTRAINT fk_universe_securities_universe_id FOREIGN KEY (universe_id) REFERENCES universes(id),
                                        CONSTRAINT uk_universe_securities_universe_security UNIQUE (universe_id, security_id),
                                        INDEX idx_universe_securities_universe_id (universe_id)
									);`); err != nil {
				return err
			}

			return nil
		},
	}
}
