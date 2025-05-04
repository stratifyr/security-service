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

			if _, err := d.SQL.Exec(`CREATE TABLE security_stats (
										id INT PRIMARY KEY AUTO_INCREMENT,
										security_id INT NOT NULL,
										date DATE NOT NULL,
										open DECIMAL(10,2) NOT NULL,
										close DECIMAL(10,2) NOT NULL,
										high DECIMAL(10,2) NOT NULL,
										low DECIMAL(10,2) NOT NULL,
										volume INT NOT NULL,
										created_at TIMESTAMP NOT NULL,
										updated_at TIMESTAMP NOT NULL,
                             
                                        CONSTRAINT uk_security_prices_security_id_date UNIQUE (security_id, date),
										CONSTRAINT fk_security_prices_security_id FOREIGN KEY (security_id) REFERENCES securities(id),
                                        INDEX idx_security_prices_security_id_date (security_id, date)
									);`); err != nil {
				return err
			}

			if _, err := d.SQL.Exec(`CREATE TABLE metrics (
										id INT PRIMARY KEY AUTO_INCREMENT,
										name varchar(50) NOT NULL,
										type INT NOT NULL,
										created_at TIMESTAMP NOT NULL,
										updated_at TIMESTAMP NOT NULL,
										
										CONSTRAINT uk_metrics_name UNIQUE (name),
                                        INDEX idx_metrics_type (type)
									);`); err != nil {
				return err
			}

			if _, err := d.SQL.Exec(`CREATE TABLE security_metrics (
										id INT PRIMARY KEY AUTO_INCREMENT,
										security_id INT NOT NULL,
										metric_id INT NOT NULL,
										date DATE NOT NULL,
										value DECIMAL(10,2) NOT NULL,
										created_at TIMESTAMP NOT NULL,
										updated_at TIMESTAMP NOT NULL,
                             
                                        CONSTRAINT uk_security_metrics_security_id_metric_id_date UNIQUE (security_id, metric_id, date),
										CONSTRAINT fk_security_metrics_security_id FOREIGN KEY (security_id) REFERENCES securities(id),
										CONSTRAINT fk_security_metrics_metric_id FOREIGN KEY (metric_id) REFERENCES metrics(id),
                                        INDEX idx_security_prices_security_id_date (security_id, date)
									);`); err != nil {
				return err
			}

			return nil
		},
	}
}
