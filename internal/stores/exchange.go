package stores

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource"
	"gofr.dev/pkg/gofr/http"
)

type ExchangeStore interface {
	List(ctx *gofr.Context, filter *ExchangeFilter, limit, offset int) ([]*Exchange, error)
	Retrieve(ctx *gofr.Context, id int) (*Exchange, error)
	Create(ctx *gofr.Context, exchange *Exchange) (*Exchange, error)
	Update(ctx *gofr.Context, id int, exchange *Exchange) (*Exchange, error)
	Delete(ctx *gofr.Context, id int) error
}

type ExchangeFilter struct {
	Name string
}

type Exchange struct {
	ID        int
	Name      string
	Currency  Currency
	ForexRate float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type exchangeStore struct{}

func NewExchangeStore() *exchangeStore {
	return &exchangeStore{}
}

func (*exchangeStore) List(ctx *gofr.Context, filter *ExchangeFilter, limit, offset int) ([]*Exchange, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, name, currency, forex_rate, created_at, updated_at
              FROM exchanges %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var exchanges []*Exchange

	for rows.Next() {
		var e Exchange

		err = rows.Scan(&e.ID, &e.Name, &e.Currency, &e.ForexRate, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		exchanges = append(exchanges, &e)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return exchanges, nil
}

func (*exchangeStore) Retrieve(ctx *gofr.Context, id int) (*Exchange, error) {
	var e Exchange

	query := `SELECT id, name, currency, forex_rate, created_at, updated_at
              FROM exchanges WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&e.ID, &e.Name, &e.Currency,
		&e.ForexRate, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, http.ErrorEntityNotFound{Name: "exchanges", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &e, nil
}

func (s *exchangeStore) Create(ctx *gofr.Context, e *Exchange) (*Exchange, error) {
	query := `INSERT INTO exchanges (name, currency, forex_rate, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?)`

	result, err := ctx.SQL.ExecContext(ctx, query, e.Name, e.Currency, e.ForexRate, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	invalidateCache(ctx, e)

	return s.Retrieve(ctx, int(id))
}

func (s *exchangeStore) Update(ctx *gofr.Context, id int, e *Exchange) (*Exchange, error) {
	query := `UPDATE exchanges SET name = ?, currency = ?, forex_rate = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, e.Name, e.Currency, e.ForexRate,
		e.CreatedAt, e.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	invalidateCache(ctx, e)

	return s.Retrieve(ctx, id)
}

func (*exchangeStore) Delete(ctx *gofr.Context, id int) error {
	_, err := ctx.SQL.ExecContext(ctx, `DELETE FROM exchanges WHERE id = ?`, id)
	if err != nil {
		return datasource.ErrorDB{Err: err}
	}

	invalidateCache(ctx, &Exchange{})

	return nil
}

func (f *ExchangeFilter) buildWhereClause() (clause string, values []any) {
	if f.Name != "" {
		clause += " AND name = ?"

		values = append(values, f.Name)
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}

func (*Exchange) cacheInvalidations() []string {
	return []string{
		SecuritiesClientCacheKey,
	}
}
