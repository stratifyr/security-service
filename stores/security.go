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

type SecurityStore interface {
	Index(ctx *gofr.Context, filter *SecurityFilter, limit, offset int) ([]*Security, error)
	Count(ctx *gofr.Context, filter *SecurityFilter) (int, error)
	Retrieve(ctx *gofr.Context, id int) (*Security, error)
	Create(ctx *gofr.Context, security *Security) (*Security, error)
	Update(ctx *gofr.Context, id int, security *Security) (*Security, error)
}

type SecurityFilter struct {
	IDs    []int
	Symbol string
}

type Security struct {
	ID        int
	ISIN      string
	Symbol    string
	Industry  Industry
	Name      string
	Image     string
	LTP       float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type securityStore struct{}

func NewSecurityStore() *securityStore {
	return &securityStore{}
}

func (s *securityStore) Index(ctx *gofr.Context, filter *SecurityFilter, limit, offset int) ([]*Security, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, isin, symbol, industry, name, image, ltp, created_at, updated_at
              FROM securities %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var securities []*Security

	for rows.Next() {
		var st Security

		err = rows.Scan(&st.ID, &st.ISIN, &st.Symbol, &st.Industry, &st.Name, &st.Image, &st.LTP, &st.CreatedAt, &st.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		securities = append(securities, &st)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return securities, nil
}

func (s *securityStore) Count(ctx *gofr.Context, filter *SecurityFilter) (int, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT COUNT(*) FROM securities %s`

	var count int

	err := ctx.SQL.QueryRowContext(ctx, fmt.Sprintf(query, whereClause), values...).Scan(&count)
	if err != nil {
		return 0, datasource.ErrorDB{Err: err}
	}

	return count, nil
}

func (s *securityStore) Retrieve(ctx *gofr.Context, id int) (*Security, error) {
	var st Security

	query := `SELECT id, isin, symbol, industry, name, image, ltp, created_at, updated_at
              FROM securities WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&st.ID, &st.ISIN, &st.Symbol, &st.Industry, &st.Name, &st.Image, &st.LTP, &st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.ErrorEntityNotFound{Name: "securities", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return &st, nil
}

func (s *securityStore) Create(ctx *gofr.Context, st *Security) (*Security, error) {
	query := "INSERT INTO securities (isin, symbol, industry, name, image, ltp, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"

	result, err := ctx.SQL.ExecContext(ctx, query, st.ISIN, st.Symbol, st.Industry, st.Name, st.Image, st.LTP, st.CreatedAt, st.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *securityStore) Update(ctx *gofr.Context, id int, st *Security) (*Security, error) {
	query := `UPDATE securities SET isin = ?, symbol = ?, industry = ?, name = ?, image = ?, last_traded_price = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err := ctx.SQL.ExecContext(ctx, query, st.ISIN, st.Symbol, st.Industry, st.Name, st.Image, st.LTP, st.CreatedAt, st.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (f *SecurityFilter) buildWhereClause() (clause string, values []interface{}) {
	if len(f.IDs) > 0 {
		var placeHolders []string

		for i := range f.IDs {
			placeHolders = append(placeHolders, "?")
			values = append(values, f.IDs[i])
		}

		clause += " AND id IN (" + strings.Join(placeHolders, ", ") + ")"
	}

	if f.Symbol != "" {
		clause += " AND symbol = ?"

		values = append(values, f.Symbol)
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
