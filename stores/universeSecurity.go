package stores

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource"
	"gofr.dev/pkg/gofr/http"
)

type UniverseSecurityStore interface {
	Index(ctx *gofr.Context, filter *UniverseSecurityFilter, limit, offset int) ([]*UniverseSecurity, error)
	Count(ctx *gofr.Context, filter *UniverseSecurityFilter) (int, error)
	Retrieve(ctx *gofr.Context, id int) (*UniverseSecurity, error)
	Create(ctx *gofr.Context, payload *UniverseSecurity) (*UniverseSecurity, error)
	Update(ctx *gofr.Context, id int, payload *UniverseSecurity) (*UniverseSecurity, error)
	Delete(ctx *gofr.Context, id int) error
}

type UniverseSecurityFilter struct {
	UniverseIDs []int
	Status      string
}

type UniverseSecurity struct {
	ID         int
	UniverseID int
	SecurityID int
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type universeSecurityStore struct{}

func NewUniverseSecurityStore() *universeSecurityStore {
	return &universeSecurityStore{}
}

func (s *universeSecurityStore) Index(ctx *gofr.Context, filter *UniverseSecurityFilter, limit, offset int) ([]*UniverseSecurity, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, universe_id, security_id, status, created_at, updated_at
              FROM universe_securities %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var universeSecurities []*UniverseSecurity

	for rows.Next() {
		var us UniverseSecurity

		err = rows.Scan(&us.ID, &us.UniverseID, &us.SecurityID, &us.Status, &us.CreatedAt, &us.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		universeSecurities = append(universeSecurities, &us)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return universeSecurities, nil
}

func (s *universeSecurityStore) Count(ctx *gofr.Context, filter *UniverseSecurityFilter) (int, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT COUNT(*) FROM universe_securities %s`

	var count int

	err := ctx.SQL.QueryRowContext(ctx, fmt.Sprintf(query, whereClause), values...).Scan(&count)
	if err != nil {
		return 0, datasource.ErrorDB{Err: err}
	}

	return count, nil
}

func (s *universeSecurityStore) Retrieve(ctx *gofr.Context, id int) (*UniverseSecurity, error) {
	var us UniverseSecurity

	query := `SELECT id, universe_id, security_id, status, created_at, updated_at
              FROM universe_securities WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&us.ID, &us.UniverseID, &us.SecurityID, &us.Status, &us.CreatedAt, &us.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.ErrorEntityNotFound{Name: "universe_security", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &us, nil
}

func (s *universeSecurityStore) Create(ctx *gofr.Context, us *UniverseSecurity) (*UniverseSecurity, error) {
	query := "INSERT INTO universe_securities (universe_id, security_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, us.UniverseID, us.SecurityID, us.Status, us.CreatedAt, us.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *universeSecurityStore) Update(ctx *gofr.Context, id int, us *UniverseSecurity) (*UniverseSecurity, error) {
	query := `UPDATE universe_securities SET universe_id = ?, security_id = ?, status = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, us.UniverseID, us.SecurityID, us.Status, us.CreatedAt, us.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (s *universeSecurityStore) Delete(ctx *gofr.Context, id int) error {
	_, err := ctx.SQL.ExecContext(ctx, `DELETE FROM universe_securities WHERE id = ?`, id)
	if err != nil {
		return datasource.ErrorDB{Err: err}
	}

	return nil
}

func (f *UniverseSecurityFilter) buildWhereClause() (clause string, values []interface{}) {
	if len(f.UniverseIDs) > 0 {
		var placeHolders []string

		for i := range f.UniverseIDs {
			placeHolders = append(placeHolders, "?")
			values = append(values, f.UniverseIDs[i])
		}

		clause += " AND universe_id IN (" + strings.Join(placeHolders, ", ") + ")"
	}

	if f.Status != "" {
		clause += " AND status = ?"

		values = append(values, f.Status)
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
