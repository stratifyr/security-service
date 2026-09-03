package stores

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource"
	"gofr.dev/pkg/gofr/http"
)

type MarketDataJobStore interface {
	Index(ctx *gofr.Context, filter *MarketDataJobFilter, limit, offset int) ([]*MarketDataJob, error)
	Count(ctx *gofr.Context, filter *MarketDataJobFilter) (int, error)
	Retrieve(ctx *gofr.Context, id int) (*MarketDataJob, error)
	Create(ctx *gofr.Context, marketHoliday *MarketDataJob) (*MarketDataJob, error)
	Update(ctx *gofr.Context, id int, marketHoliday *MarketDataJob) (*MarketDataJob, error)
}

type MarketDataJobFilter struct {
	Status string
}

type MarketDataJob struct {
	ID        int
	Type      MarketDataJobType
	Status    string
	Logs      *json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type marketDataJobStore struct{}

func NewMarketDataJobStore() *marketDataJobStore {
	return &marketDataJobStore{}
}

func (*marketDataJobStore) Index(ctx *gofr.Context, filter *MarketDataJobFilter, limit, offset int) ([]*MarketDataJob, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, type, status, logs, created_at, updated_at
              FROM market_data_jobs %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var marketDataJobs []*MarketDataJob

	for rows.Next() {
		var mdj MarketDataJob

		err = rows.Scan(&mdj.ID, &mdj.Type, &mdj.Status, &mdj.Logs, &mdj.CreatedAt, &mdj.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		marketDataJobs = append(marketDataJobs, &mdj)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return marketDataJobs, nil
}

func (*marketDataJobStore) Count(ctx *gofr.Context, filter *MarketDataJobFilter) (int, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT COUNT(*) FROM market_data_jobs %s`

	var count int

	err := ctx.SQL.QueryRowContext(ctx, fmt.Sprintf(query, whereClause), values...).Scan(&count)
	if err != nil {
		return 0, datasource.ErrorDB{Err: err}
	}

	return count, nil
}

func (*marketDataJobStore) Retrieve(ctx *gofr.Context, id int) (*MarketDataJob, error) {
	var mdj MarketDataJob

	query := `SELECT id, type, status, logs, created_at, updated_at
              FROM market_data_jobs WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&mdj.ID, &mdj.Type, &mdj.Status, &mdj.Logs, &mdj.CreatedAt, &mdj.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, http.ErrorEntityNotFound{Name: "market-data-jobs", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &mdj, nil
}

func (s *marketDataJobStore) Create(ctx *gofr.Context, mdj *MarketDataJob) (*MarketDataJob, error) {
	query := "INSERT INTO market_data_jobs (type, status, logs, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, mdj.Type, mdj.Status, mdj.Logs, mdj.CreatedAt, mdj.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *marketDataJobStore) Update(ctx *gofr.Context, id int, mdj *MarketDataJob) (*MarketDataJob, error) {
	query := `UPDATE market_data_jobs SET type = ?, status = ?, logs = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, mdj.Type, mdj.Status, mdj.Logs, mdj.CreatedAt, mdj.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (f *MarketDataJobFilter) buildWhereClause() (clause string, values []any) {
	if f.Status != "" {
		clause += " AND status = ?"

		values = append(values, f.Status)
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
