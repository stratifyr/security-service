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

type SecurityMetricStore interface {
	Index(ctx *gofr.Context, filter *SecurityMetricFilter, limit, offset int) ([]*SecurityMetric, error)
	Count(ctx *gofr.Context, filter *SecurityMetricFilter) (int, error)
	Retrieve(ctx *gofr.Context, id int) (*SecurityMetric, error)
	Create(ctx *gofr.Context, sm *SecurityMetric) (*SecurityMetric, error)
	Update(ctx *gofr.Context, id int, sm *SecurityMetric) (*SecurityMetric, error)
}

type SecurityMetricFilter struct {
	SecurityID int
	MetricID   int
	Date       time.Time
}

type SecurityMetric struct {
	ID         int
	SecurityID int
	MetricID   int
	Date       time.Time
	Value      float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type securityMetricStore struct{}

func NewSecurityMetricStore() *securityMetricStore {
	return &securityMetricStore{}
}

func (s *securityMetricStore) Index(ctx *gofr.Context, filter *SecurityMetricFilter, limit, offset int) ([]*SecurityMetric, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, security_id, metric_id, date, value, created_at, updated_at
              FROM security_metrics %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var securityMetrics []*SecurityMetric

	for rows.Next() {
		var sm SecurityMetric

		err = rows.Scan(&sm.ID, &sm.SecurityID, &sm.MetricID, &sm.Date, &sm.Value, &sm.CreatedAt, &sm.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		securityMetrics = append(securityMetrics, &sm)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return securityMetrics, nil
}

func (s *securityMetricStore) Count(ctx *gofr.Context, filter *SecurityMetricFilter) (int, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT COUNT(*) FROM security_metrics %s`

	var count int

	err := ctx.SQL.QueryRowContext(ctx, fmt.Sprintf(query, whereClause), values...).Scan(&count)
	if err != nil {
		return 0, datasource.ErrorDB{Err: err}
	}

	return count, nil
}

func (s *securityMetricStore) Retrieve(ctx *gofr.Context, id int) (*SecurityMetric, error) {
	var sm SecurityMetric

	query := `SELECT id, security_id, metric_id, date, value, created_at, updated_at
              FROM security_metrics WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&sm.ID, &sm.SecurityID, &sm.MetricID, &sm.Date, &sm.Value, &sm.CreatedAt, &sm.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.ErrorEntityNotFound{Name: "security-metrics", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &sm, nil
}

func (s *securityMetricStore) Create(ctx *gofr.Context, sm *SecurityMetric) (*SecurityMetric, error) {
	query := "INSERT INTO security_metrics (security_id, metric_id, date, value, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, sm.SecurityID, sm.MetricID, sm.Date, sm.Value, sm.CreatedAt, sm.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *securityMetricStore) Update(ctx *gofr.Context, id int, sm *SecurityMetric) (*SecurityMetric, error) {
	query := `UPDATE security_metrics SET security_id = ?, metric_id = ?, date = ?, value = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, sm.SecurityID, sm.MetricID, sm.Date, sm.Value, sm.CreatedAt, sm.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (f *SecurityMetricFilter) buildWhereClause() (clause string, values []interface{}) {
	if f.SecurityID != 0 {
		clause += " AND security_id = ?"

		values = append(values, f.SecurityID)
	}

	if f.MetricID != 0 {
		clause += " AND metric_id = ?"

		values = append(values, f.MetricID)
	}

	if f.Date != (time.Time{}) {
		clause += " AND date = ?"

		values = append(values, f.Date.Format(time.DateOnly))
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
