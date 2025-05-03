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

type MetricStore interface {
	Index(ctx *gofr.Context, filter *MetricFilter, limit, offset int) ([]*Metric, error)
	Count(ctx *gofr.Context, filter *MetricFilter) (int, error)
	Retrieve(ctx *gofr.Context, id int) (*Metric, error)
	Create(ctx *gofr.Context, metric *Metric) (*Metric, error)
	Update(ctx *gofr.Context, id int, metric *Metric) (*Metric, error)
}

type MetricFilter struct {
	Type *MetricType
}

type Metric struct {
	ID        int
	Name      string
	Type      MetricType
	CreatedAt time.Time
	UpdatedAt time.Time
}

type metricStore struct{}

func NewMetricStore() *metricStore {
	return &metricStore{}
}

func (s *metricStore) Index(ctx *gofr.Context, filter *MetricFilter, limit, offset int) ([]*Metric, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, name, type, created_at, updated_at
              FROM metrics %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var metrics []*Metric

	for rows.Next() {
		var m Metric

		err = rows.Scan(&m.ID, &m.Name, &m.Type, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		metrics = append(metrics, &m)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return metrics, nil
}

func (s *metricStore) Count(ctx *gofr.Context, filter *MetricFilter) (int, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT COUNT(*) FROM metrics %s`

	var count int

	err := ctx.SQL.QueryRowContext(ctx, fmt.Sprintf(query, whereClause), values...).Scan(&count)
	if err != nil {
		return 0, datasource.ErrorDB{Err: err}
	}

	return count, nil
}

func (s *metricStore) Retrieve(ctx *gofr.Context, id int) (*Metric, error) {
	var m Metric

	query := `SELECT id, name, type, created_at, updated_at
              FROM metrics WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.Name, &m.Type, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.ErrorEntityNotFound{Name: "metrics", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &m, nil
}

func (s *metricStore) Create(ctx *gofr.Context, m *Metric) (*Metric, error) {
	query := "INSERT INTO metrics (name, type, created_at, updated_at) VALUES (?, ?, ?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, m.Name, m.Type, m.CreatedAt, m.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *metricStore) Update(ctx *gofr.Context, id int, m *Metric) (*Metric, error) {
	query := `UPDATE metrics SET name = ?, type = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, m.Name, m.Type, m.CreatedAt, m.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (f *MetricFilter) buildWhereClause() (clause string, values []interface{}) {
	if f.Type != nil {
		clause += " AND type = ?"

		values = append(values, f.Type)
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
