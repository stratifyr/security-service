package stores

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource"
	"gofr.dev/pkg/gofr/http"
)

type MetricStore interface {
	Index(ctx *gofr.Context, filter *MetricFilter) ([]*Metric, error)
	Retrieve(ctx *gofr.Context, id int) (*Metric, error)
	Create(ctx *gofr.Context, metric *Metric) (*Metric, error)
	Update(ctx *gofr.Context, id int, metric *Metric) (*Metric, error)
}

type MetricFilter struct {
	MaxTier *int
}

type Metric struct {
	ID        int
	Name      string
	Type      MetricType
	Period    int
	Indicator MetricIndicator
	Tier      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type metricStore struct {
	cache *sync.Map
}

func NewMetricStore() *metricStore {
	return &metricStore{cache: new(sync.Map)}
}

func (s *metricStore) Index(ctx *gofr.Context, filter *MetricFilter) ([]*Metric, error) {
	cache, ok := s.cache.Load(filter.getCacheKey())
	if ok {
		return cache.([]*Metric), nil
	}

	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, name, type, period, indicator, tier, created_at, updated_at
              FROM metrics %s`

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var metrics []*Metric

	for rows.Next() {
		var m Metric

		err = rows.Scan(&m.ID, &m.Name, &m.Type, &m.Period, &m.Indicator, &m.Tier, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		metrics = append(metrics, &m)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	s.cache.Store(filter.getCacheKey(), metrics)

	return metrics, nil
}

func (s *metricStore) Retrieve(ctx *gofr.Context, id int) (*Metric, error) {
	var m Metric

	query := `SELECT id, name, type, period, indicator, tier, created_at, updated_at
              FROM metrics WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.Name, &m.Type, &m.Period, &m.Indicator, &m.Tier, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.ErrorEntityNotFound{Name: "metrics", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &m, nil
}

func (s *metricStore) Create(ctx *gofr.Context, m *Metric) (*Metric, error) {
	s.cache.Clear()

	query := "INSERT INTO metrics (name, type, period, indicator, tier, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, m.Name, m.Type, m.Period, m.Indicator, m.Tier, m.CreatedAt, m.UpdatedAt)
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
	s.cache.Clear()

	query := `UPDATE metrics SET name = ?, type = ?, period = ?, indicator = ?, tier = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, m.Name, m.Type, m.Period, m.Indicator, m.Tier, m.CreatedAt, m.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (f *MetricFilter) buildWhereClause() (clause string, values []interface{}) {
	if f.MaxTier != nil {
		clause += " AND tier <= ?"

		values = append(values, *f.MaxTier)
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}

func (f *MetricFilter) getCacheKey() string {
	return fmt.Sprintf("metrics:max_tier:%d", f.MaxTier)
}
