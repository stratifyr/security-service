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

type SecurityStatStore interface {
	Index(ctx *gofr.Context, filter *SecurityStatFilter, limit, offset int) ([]*SecurityStat, error)
	Count(ctx *gofr.Context, filter *SecurityStatFilter) (int, error)
	Retrieve(ctx *gofr.Context, id int) (*SecurityStat, error)
	Create(ctx *gofr.Context, ss *SecurityStat) (*SecurityStat, error)
	Update(ctx *gofr.Context, id int, ss *SecurityStat) (*SecurityStat, error)
}

type SecurityStatFilter struct {
	SecurityID int
	Dates      []time.Time
}

type SecurityStat struct {
	ID         int
	SecurityID int
	Date       time.Time
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Volume     int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type securityStatStore struct{}

func NewSecurityStatStore() *securityStatStore {
	return &securityStatStore{}
}

func (s *securityStatStore) Index(ctx *gofr.Context, filter *SecurityStatFilter, limit, offset int) ([]*SecurityStat, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, security_id, date, open, close, high, low, volume, created_at, updated_at
              FROM security_stats %s
              ORDER BY date DESC`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var securityStats []*SecurityStat

	for rows.Next() {
		var ss SecurityStat

		err = rows.Scan(&ss.ID, &ss.SecurityID, &ss.Date, &ss.Open, &ss.Close, &ss.High, &ss.Low, &ss.Volume, &ss.CreatedAt, &ss.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		securityStats = append(securityStats, &ss)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return securityStats, nil
}

func (s *securityStatStore) Count(ctx *gofr.Context, filter *SecurityStatFilter) (int, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT COUNT(*) FROM security_stats %s`

	var count int

	err := ctx.SQL.QueryRowContext(ctx, fmt.Sprintf(query, whereClause), values...).Scan(&count)
	if err != nil {
		return 0, datasource.ErrorDB{Err: err}
	}

	return count, nil
}

func (s *securityStatStore) Retrieve(ctx *gofr.Context, id int) (*SecurityStat, error) {
	var ss SecurityStat

	query := `SELECT id, security_id, date, open, close, high, low, volume, created_at, updated_at
              FROM security_stats WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&ss.ID, &ss.SecurityID, &ss.Date, &ss.Open, &ss.Close, &ss.High, &ss.Low, &ss.Volume, &ss.CreatedAt, &ss.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.ErrorEntityNotFound{Name: "security-stats", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &ss, nil
}

func (s *securityStatStore) Create(ctx *gofr.Context, ss *SecurityStat) (*SecurityStat, error) {
	query := "INSERT INTO security_stats (security_id, date, open, close, high, low, volume, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, ss.SecurityID, ss.Date, ss.Open, ss.Close, ss.High, ss.Low, ss.Volume, ss.CreatedAt, ss.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *securityStatStore) Update(ctx *gofr.Context, id int, ss *SecurityStat) (*SecurityStat, error) {
	query := `UPDATE security_stats SET security_id = ?, date = ?, open = ?, close = ?, high = ?, low = ?, volume = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, ss.SecurityID, ss.Date, ss.Open, ss.Close, ss.High, ss.Low, ss.Volume, ss.CreatedAt, ss.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (f *SecurityStatFilter) buildWhereClause() (clause string, values []interface{}) {
	if f.SecurityID != 0 {
		clause += " AND security_id = ?"

		values = append(values, f.SecurityID)
	}

	if len(f.Dates) > 0 {
		var placeHolders []string

		for i := range f.Dates {
			placeHolders = append(placeHolders, "?")
			values = append(values, f.Dates[i].Format(time.DateOnly))
		}

		clause += " AND date IN (" + strings.Join(placeHolders, ", ") + ")"
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
