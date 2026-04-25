package stores

import (
	"database/sql"
	"strconv"
	"sync"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource"
	"gofr.dev/pkg/gofr/http"
)

const metricsCacheKey = "cache:security-service:store:metrics"

type MetricStore interface {
	Index(ctx *gofr.Context) ([]*Metric, error)
	Retrieve(ctx *gofr.Context, id int) (*Metric, error)
	Create(ctx *gofr.Context, metric *Metric) (*Metric, error)
	Update(ctx *gofr.Context, id int, metric *Metric) (*Metric, error)
}

type Metric struct {
	ID        int
	Name      string
	Type      MetricType
	Period    int
	Indicator MetricIndicator
	CreatedAt time.Time
	UpdatedAt time.Time
}

type metricStore struct {
	cache *sync.Map
}

func NewMetricStore() *metricStore {
	return &metricStore{cache: new(sync.Map)}
}

func (s *metricStore) Index(ctx *gofr.Context) ([]*Metric, error) {
	cache, ok := s.cache.Load(metricsCacheKey)
	if ok {
		return cache.([]*Metric), nil
	}

	query := `SELECT id, name, type, period, indicator, created_at, updated_at
              FROM metrics`

	rows, err := ctx.SQL.QueryContext(ctx, query)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var metrics []*Metric

	for rows.Next() {
		var m Metric

		err = rows.Scan(&m.ID, &m.Name, &m.Type, &m.Period, &m.Indicator, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		metrics = append(metrics, &m)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	s.cache.Store(metricsCacheKey, metrics)

	return metrics, nil
}

func (s *metricStore) Retrieve(ctx *gofr.Context, id int) (*Metric, error) {
	var m Metric

	query := `SELECT id, name, type, period, indicator, created_at, updated_at
              FROM metrics WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.Name, &m.Type, &m.Period, &m.Indicator, &m.CreatedAt, &m.UpdatedAt)
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

	query := "INSERT INTO metrics (name, type, period, indicator, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, m.Name, m.Type, m.Period, m.Indicator, m.CreatedAt, m.UpdatedAt)
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

	query := `UPDATE metrics SET name = ?, type = ?, period = ?, indicator = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, m.Name, m.Type, m.Period, m.Indicator, m.CreatedAt, m.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}
