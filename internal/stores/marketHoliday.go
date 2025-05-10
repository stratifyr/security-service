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

type MarketHolidayStore interface {
	Index(ctx *gofr.Context, filter *MarketHolidayFilter, limit, offset int) ([]*MarketHoliday, error)
	Count(ctx *gofr.Context, filter *MarketHolidayFilter) (int, error)
	Retrieve(ctx *gofr.Context, id int) (*MarketHoliday, error)
	Create(ctx *gofr.Context, marketHoliday *MarketHoliday) (*MarketHoliday, error)
	Update(ctx *gofr.Context, id int, marketHoliday *MarketHoliday) (*MarketHoliday, error)
	Delete(ctx *gofr.Context, id int) error
}

type MarketHolidayFilter struct {
	Date        time.Time
	DateBetween *struct {
		StartDate time.Time
		EndDate   time.Time
	}
}

type MarketHoliday struct {
	ID          int
	Date        time.Time
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type marketHolidayStore struct{}

func NewMarketHolidayStore() *marketHolidayStore {
	return &marketHolidayStore{}
}

func (s *marketHolidayStore) Index(ctx *gofr.Context, filter *MarketHolidayFilter, limit, offset int) ([]*MarketHoliday, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, date, description, created_at, updated_at
              FROM market_holidays %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var marketHolidays []*MarketHoliday

	for rows.Next() {
		var mh MarketHoliday

		err = rows.Scan(&mh.ID, &mh.Date, &mh.Description, &mh.CreatedAt, &mh.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		marketHolidays = append(marketHolidays, &mh)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return marketHolidays, nil
}

func (s *marketHolidayStore) Count(ctx *gofr.Context, filter *MarketHolidayFilter) (int, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT COUNT(*) FROM market_holidays %s`

	var count int

	err := ctx.SQL.QueryRowContext(ctx, fmt.Sprintf(query, whereClause), values...).Scan(&count)
	if err != nil {
		return 0, datasource.ErrorDB{Err: err}
	}

	return count, nil
}

func (s *marketHolidayStore) Retrieve(ctx *gofr.Context, id int) (*MarketHoliday, error) {
	var mh MarketHoliday

	query := `SELECT id, date, description, created_at, updated_at
              FROM market_holidays WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&mh.ID, &mh.Date, &mh.Description, &mh.CreatedAt, &mh.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.ErrorEntityNotFound{Name: "market-holidays", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &mh, nil
}

func (s *marketHolidayStore) Create(ctx *gofr.Context, mh *MarketHoliday) (*MarketHoliday, error) {
	query := "INSERT INTO market_holidays (date, description, created_at, updated_at) VALUES (?, ?, ?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, mh.Date, mh.Description, mh.CreatedAt, mh.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *marketHolidayStore) Update(ctx *gofr.Context, id int, mh *MarketHoliday) (*MarketHoliday, error) {
	query := `UPDATE market_holidays SET date = ?, description = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, mh.Date, mh.Description, mh.CreatedAt, mh.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (s *marketHolidayStore) Delete(ctx *gofr.Context, id int) error {
	_, err := ctx.SQL.ExecContext(ctx, `DELETE FROM market_holidays WHERE id = ?`, id)
	if err != nil {
		return datasource.ErrorDB{Err: err}
	}

	return nil
}

func (f *MarketHolidayFilter) buildWhereClause() (clause string, values []interface{}) {
	if f.Date != (time.Time{}) {
		clause += " AND date = ?"

		values = append(values, f.Date.Format(time.DateOnly))
	}

	if f.DateBetween != nil {
		clause += " AND date BETWEEN ? AND ?"

		values = append(values, f.DateBetween.StartDate.Format(time.DateOnly), f.DateBetween.EndDate.Format(time.DateOnly))
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
