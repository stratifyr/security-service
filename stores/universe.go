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

type UniverseStore interface {
	Index(ctx *gofr.Context, filter *UniverseFilter, limit, offset int) ([]*Universe, error)
	Count(ctx *gofr.Context, filter *UniverseFilter) (int, error)
	Retrieve(ctx *gofr.Context, id int) (*Universe, error)
	Create(ctx *gofr.Context, universe *Universe) (*Universe, error)
	Update(ctx *gofr.Context, id int, universe *Universe) (*Universe, error)
}

type UniverseFilter struct {
	UserIDs []int
}

type Universe struct {
	ID                 int
	UserID             int
	Name               string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	UniverseSecurities []*UniverseSecurity
}

type universeStore struct {
	universeSecurityStore UniverseSecurityStore
}

func NewUniverseStore(universeSecurityStore UniverseSecurityStore) *universeStore {
	return &universeStore{universeSecurityStore: universeSecurityStore}
}

func (s *universeStore) Index(ctx *gofr.Context, filter *UniverseFilter, limit, offset int) ([]*Universe, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, user_id, name, created_at, updated_at
              FROM universes %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var universes []*Universe

	for rows.Next() {
		var u Universe

		err = rows.Scan(&u.ID, &u.UserID, &u.Name, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		universes = append(universes, &u)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	for _, u := range universes {
		u.UniverseSecurities, err = s.universeSecurityStore.Index(ctx, &UniverseSecurityFilter{UniverseIDs: []int{u.ID}}, 0, 0)
		if err != nil {
			return nil, err
		}
	}

	return universes, nil
}

func (s *universeStore) Count(ctx *gofr.Context, filter *UniverseFilter) (int, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT COUNT(*) FROM universes %s`

	var count int

	err := ctx.SQL.QueryRowContext(ctx, fmt.Sprintf(query, whereClause), values...).Scan(&count)
	if err != nil {
		return 0, datasource.ErrorDB{Err: err}
	}

	return count, nil
}

func (s *universeStore) Retrieve(ctx *gofr.Context, id int) (*Universe, error) {
	var u Universe

	query := `SELECT id, user_id, name, created_at, updated_at
              FROM universes WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.UserID, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.ErrorEntityNotFound{Name: "securities", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	u.UniverseSecurities, err = s.universeSecurityStore.Index(ctx, &UniverseSecurityFilter{UniverseIDs: []int{id}}, 0, 0)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *universeStore) Create(ctx *gofr.Context, u *Universe) (*Universe, error) {
	query := "INSERT INTO universes (user_id, name, created_at, updated_at) VALUES (?, ?, ?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, u.UserID, u.Name, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	for i := range u.UniverseSecurities {
		u.UniverseSecurities[i].UniverseID = int(id)

		if _, err = s.universeSecurityStore.Create(ctx, u.UniverseSecurities[i]); err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *universeStore) Update(ctx *gofr.Context, id int, u *Universe) (*Universe, error) {
	query := `UPDATE universes SET user_id = ?, name = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, u.UserID, u.Name, u.CreatedAt, u.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	for i := range u.UniverseSecurities {
		if u.UniverseSecurities[i].ID != 0 {
			if _, err = s.universeSecurityStore.Create(ctx, u.UniverseSecurities[i]); err != nil {
				return nil, datasource.ErrorDB{Err: err}
			}

			continue
		}

		if _, err = s.universeSecurityStore.Update(ctx, u.UniverseSecurities[i].ID, u.UniverseSecurities[i]); err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}
	}

	return s.Retrieve(ctx, id)
}

func (f *UniverseFilter) buildWhereClause() (clause string, values []interface{}) {
	if len(f.UserIDs) > 0 {
		var placeHolders []string

		for i := range f.UserIDs {
			placeHolders = append(placeHolders, "?")
			values = append(values, f.UserIDs[i])
		}

		clause += " AND id IN (" + strings.Join(placeHolders, ", ") + ")"
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
