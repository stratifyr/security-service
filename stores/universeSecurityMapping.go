package stores

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource"
	"gofr.dev/pkg/gofr/http"
)

type UniverseSecurityMappingStore interface {
	Index(ctx *gofr.Context, filter *UniverseSecurityMappingFilter, limit, offset int) ([]*UniverseSecurityMapping, error)
	Count(ctx *gofr.Context, filter *UniverseSecurityMappingFilter) (int, error)
	Retrieve(ctx *gofr.Context, id int) (*UniverseSecurityMapping, error)
	Create(ctx *gofr.Context, mapping *UniverseSecurityMapping) (*UniverseSecurityMapping, error)
	Update(ctx *gofr.Context, id int, mapping *UniverseSecurityMapping) (*UniverseSecurityMapping, error)
	Delete(ctx *gofr.Context, id int) error
}

type UniverseSecurityMapping struct {
	ID         int
	UniverseID int
	SecurityID int
}

type UniverseSecurityMappingFilter struct {
	UniverseID int
}

type universeSecurityMappingStore struct{}

func NewUniverseSecurityMappingStore() *universeSecurityMappingStore {
	return &universeSecurityMappingStore{}
}

func (s *universeSecurityMappingStore) Index(ctx *gofr.Context, filter *UniverseSecurityMappingFilter, limit, offset int) ([]*UniverseSecurityMapping, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, universe_id, security_id
              FROM universe_security_mapping %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var universeSecurityMappings []*UniverseSecurityMapping

	for rows.Next() {
		var usm UniverseSecurityMapping

		err = rows.Scan(&usm.ID, &usm.UniverseID, &usm.SecurityID)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		universeSecurityMappings = append(universeSecurityMappings, &usm)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return universeSecurityMappings, nil
}

func (s *universeSecurityMappingStore) Count(ctx *gofr.Context, filter *SecurityFilter) (int, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT COUNT(*) FROM universe_security_mapping %s`

	var count int

	err := ctx.SQL.QueryRowContext(ctx, fmt.Sprintf(query, whereClause), values...).Scan(&count)
	if err != nil {
		return 0, datasource.ErrorDB{Err: err}
	}

	return count, nil
}

func (s *universeSecurityMappingStore) Retrieve(ctx *gofr.Context, id int) (*UniverseSecurityMapping, error) {
	var usm UniverseSecurityMapping

	query := `SELECT id, universe_id, security_id
              FROM universe_security_mapping WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&usm.ID, &usm.UniverseID, &usm.SecurityID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.ErrorEntityNotFound{Name: "universe_security_mappings", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &usm, nil
}

func (s *universeSecurityMappingStore) Create(ctx *gofr.Context, usm *UniverseSecurityMapping) (*UniverseSecurityMapping, error) {
	query := "INSERT INTO universe_security_mapping (universe_id, security_id) VALUES (?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, usm.UniverseID, usm.SecurityID)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *universeSecurityMappingStore) Update(ctx *gofr.Context, id int, usm *UniverseSecurityMapping) (*UniverseSecurityMapping, error) {
	query := `UPDATE universe_security_mapping SET universe_id = ?, security_id = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, usm.UniverseID, usm.SecurityID, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (s *universeSecurityMappingStore) Delete(ctx *gofr.Context, id int) error {
	_, err := ctx.SQL.ExecContext(ctx, `DELETE FROM universe_security_mapping WHERE id = ?`, id)
	if err != nil {
		return datasource.ErrorDB{Err: err}
	}

	return nil
}

func (f *UniverseSecurityMappingFilter) buildWhereClause() (clause string, values []interface{}) {
	if f.UniverseID != 0 {
		clause += " AND universe_id = ?"

		values = append(values, f.UniverseID)
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
