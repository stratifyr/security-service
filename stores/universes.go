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

type Universe struct {
	ID        int
	UserID    sql.NullInt64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time

	SecurityIDs []int
}

type UniverseFilter struct {
	UserID int
}

type universeStore struct {
	universeSecurityMappingStore UniverseSecurityMappingStore
}

func NewUniverseStore(universeSecurityMappingStore UniverseSecurityMappingStore) *universeStore {
	return &universeStore{universeSecurityMappingStore: universeSecurityMappingStore}
}

func (s *universeStore) Index(ctx *gofr.Context, filter *SecurityFilter, limit, offset int) ([]*Universe, error) {
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

	for _, universe := range universes {
		mappings, err := s.universeSecurityMappingStore.Index(ctx, &UniverseSecurityMappingFilter{UniverseID: universe.ID}, 0, 0)
		if err != nil {
			return nil, err
		}

		if len(mappings) == 0 {
			continue
		}

		universe.SecurityIDs = make([]int, len(mappings))

		for i := range mappings {
			universe.SecurityIDs[i] = mappings[i].SecurityID
		}
	}

	return universes, nil
}

func (s *universeStore) Count(ctx *gofr.Context, filter *SecurityFilter) (int, error) {
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

	mappings, err := s.universeSecurityMappingStore.Index(ctx, &UniverseSecurityMappingFilter{UniverseID: id}, 0, 0)
	if err != nil {
		return nil, err
	}

	if len(mappings) == 0 {
		return &u, nil
	}

	u.SecurityIDs = make([]int, len(mappings))

	for i := range mappings {
		u.SecurityIDs[i] = mappings[i].SecurityID
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

	for _, securityID := range u.SecurityIDs {
		_, err := s.universeSecurityMappingStore.Create(ctx, &UniverseSecurityMapping{
			UniverseID: int(id),
			SecurityID: securityID})
		if err != nil {
			return nil, err
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

	mappings, err := s.universeSecurityMappingStore.Index(ctx, &UniverseSecurityMappingFilter{UniverseID: id}, 0, 0)
	if err != nil {
		return nil, err
	}

	var mappingsMap = make(map[int]*UniverseSecurityMapping)

	for i := range mappings {
		mappingsMap[mappings[i].SecurityID] = mappings[i]
	}

	for _, securityID := range u.SecurityIDs {
		if _, exists := mappingsMap[securityID]; exists {
			_, err := s.universeSecurityMappingStore.Create(ctx, &UniverseSecurityMapping{
				UniverseID: id,
				SecurityID: securityID})
			if err != nil {
				return nil, err
			}
		} else {
			delete(mappingsMap, securityID)
		}
	}

	for _, mapping := range mappingsMap {
		err := s.universeSecurityMappingStore.Delete(ctx, mapping.ID)
		if err != nil {
			return nil, err
		}
	}

	return s.Retrieve(ctx, id)
}

func (f *UniverseFilter) buildWhereClause() (clause string, values []interface{}) {
	if f.UserID != 0 {
		clause += " AND user_id = ?"

		values = append(values, f.UserID)
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
