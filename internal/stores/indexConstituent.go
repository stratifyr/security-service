package stores

import (
	"fmt"
	"strings"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource"
	"gofr.dev/pkg/gofr/datasource/sql"
)

type IndexConstituentStore interface {
	List(ctx *gofr.Context, filter *IndexConstituentFilter, limit, offset int) ([]*IndexConstituent, error)
	CreateAll(ctx *gofr.Context, txn *sql.Tx, constituents []*IndexConstituent) error
	Delete(ctx *gofr.Context, txn *sql.Tx, indexID int) error
}

type IndexConstituentFilter struct {
	IndexID int
}

type IndexConstituent struct {
	ID         int
	IndexID    int
	SecurityID int
}

type indexConstituentStore struct{}

func NewIndexConstituentStore() *indexConstituentStore {
	return &indexConstituentStore{}
}

func (*indexConstituentStore) List(ctx *gofr.Context, filter *IndexConstituentFilter, limit, offset int) ([]*IndexConstituent, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, index_id, security_id
              FROM index_constituents %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var constituents []*IndexConstituent

	for rows.Next() {
		var c IndexConstituent

		err = rows.Scan(&c.ID, &c.IndexID, &c.SecurityID)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		constituents = append(constituents, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return constituents, nil
}

func (*indexConstituentStore) CreateAll(ctx *gofr.Context, txn *sql.Tx, constituents []*IndexConstituent) error {
	if len(constituents) == 0 {
		return nil
	}

	query := `INSERT INTO index_constituents (index_id, security_id)
		      VALUES`

	values := make([]any, 0, len(constituents)*2) //nolint:mnd // expected 2 placeholders
	placeholders := make([]string, 0, len(constituents))

	for _, c := range constituents {
		placeholders = append(placeholders, "(?, ?)")
		values = append(values, c.IndexID, c.SecurityID)
	}

	query += strings.Join(placeholders, ", ")

	if _, err := txn.ExecContext(ctx, query, values...); err != nil {
		return datasource.ErrorDB{Err: err}
	}

	return nil
}

func (*indexConstituentStore) Delete(ctx *gofr.Context, txn *sql.Tx, indexID int) error {
	query := `DELETE FROM index_constituents WHERE index_id = ?`

	_, err := txn.ExecContext(ctx, query, indexID)
	if err != nil {
		return datasource.ErrorDB{Err: err}
	}

	return nil
}

func (f *IndexConstituentFilter) buildWhereClause() (clause string, values []any) {
	if f.IndexID != 0 {
		clause += " AND index_id = ?"

		values = append(values, f.IndexID)
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
