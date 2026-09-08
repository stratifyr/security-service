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

type IndexStatStore interface {
	List(ctx *gofr.Context, filter *IndexStatFilter, limit, offset int) ([]*IndexStat, error)
	Count(ctx *gofr.Context, filter *IndexStatFilter) (int, error)
	Retrieve(ctx *gofr.Context, id int) (*IndexStat, error)
	Create(ctx *gofr.Context, is *IndexStat) (*IndexStat, error)
	Update(ctx *gofr.Context, id int, is *IndexStat) (*IndexStat, error)
}

type IndexStatFilter struct {
	IndexIDs    []int
	Date        time.Time
	DateBetween *struct {
		Start time.Time
		End   time.Time
	}
}

type IndexStat struct {
	ID        int
	IndexID   int
	Date      time.Time
	Open      float64
	Close     float64
	High      float64
	Low       float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type indexStatStore struct{}

func NewIndexStatStore() *indexStatStore {
	return &indexStatStore{}
}

func (*indexStatStore) List(ctx *gofr.Context, filter *IndexStatFilter, limit, offset int) ([]*IndexStat, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, index_id, date, open, close, high, low, created_at, updated_at
              FROM index_stats %s
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

	var indexStats []*IndexStat

	for rows.Next() {
		var is IndexStat

		err = rows.Scan(&is.ID, &is.IndexID, &is.Date, &is.Open, &is.Close, &is.High, &is.Low, &is.CreatedAt, &is.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		indexStats = append(indexStats, &is)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return indexStats, nil
}

func (*indexStatStore) Count(ctx *gofr.Context, filter *IndexStatFilter) (int, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT COUNT(*) FROM index_stats %s`

	var count int

	err := ctx.SQL.QueryRowContext(ctx, fmt.Sprintf(query, whereClause), values...).Scan(&count)
	if err != nil {
		return 0, datasource.ErrorDB{Err: err}
	}

	return count, nil
}

func (*indexStatStore) Retrieve(ctx *gofr.Context, id int) (*IndexStat, error) {
	var is IndexStat

	query := `SELECT id, index_id, date, open, close, high, low, created_at, updated_at
              FROM index_stats WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&is.ID, &is.IndexID, &is.Date, &is.Open,
		&is.Close, &is.High, &is.Low, &is.CreatedAt, &is.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, http.ErrorEntityNotFound{Name: "index-stats", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &is, nil
}

func (s *indexStatStore) Create(ctx *gofr.Context, is *IndexStat) (*IndexStat, error) {
	query := `INSERT INTO index_stats (index_id, date, open, close, high, low, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := ctx.SQL.ExecContext(ctx, query, is.IndexID, is.Date, is.Open,
		is.Close, is.High, is.Low, is.CreatedAt, is.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *indexStatStore) Update(ctx *gofr.Context, id int, is *IndexStat) (*IndexStat, error) {
	query := `UPDATE index_stats SET index_id = ?, date = ?, open = ?, close = ?, 
                          high = ?, low = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, is.IndexID, is.Date, is.Open, is.Close,
		is.High, is.Low, is.CreatedAt, is.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (f *IndexStatFilter) buildWhereClause() (clause string, values []any) {
	if len(f.IndexIDs) > 0 {
		placeHolders := make([]string, 0, len(f.IndexIDs))

		for i := range f.IndexIDs {
			placeHolders = append(placeHolders, "?")
			values = append(values, f.IndexIDs[i])
		}

		clause += " AND index_id IN (" + strings.Join(placeHolders, ", ") + ")"
	}

	if f.Date != (time.Time{}) {
		clause += " AND date = ?"

		values = append(values, f.Date.Format(time.DateOnly))
	}

	if f.DateBetween != nil {
		clause += " AND date BETWEEN ? AND ?"

		values = append(values, f.DateBetween.Start.Format(time.DateOnly), f.DateBetween.End.Format(time.DateOnly))
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
