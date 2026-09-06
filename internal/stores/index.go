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
	gofrSQL "gofr.dev/pkg/gofr/datasource/sql"
	"gofr.dev/pkg/gofr/http"
)

type IndexStore interface {
	List(ctx *gofr.Context, filter *IndexFilter, limit, offset int) ([]*Index, error)
	Retrieve(ctx *gofr.Context, id int) (*Index, error)
	Create(ctx *gofr.Context, index *Index) (*Index, error)
	Update(ctx *gofr.Context, id int, index *Index) (*Index, error)
	Delete(ctx *gofr.Context, id int) error
}

type IndexFilter struct {
	Name string
}

type Index struct {
	ID           int
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Constituents []*IndexConstituent
}

type indexStore struct {
	indexConstituentStore IndexConstituentStore
}

func NewIndexStore(indexConstituentStore IndexConstituentStore) *indexStore {
	return &indexStore{indexConstituentStore: indexConstituentStore}
}

func (s *indexStore) List(ctx *gofr.Context, filter *IndexFilter, limit, offset int) ([]*Index, error) {
	whereClause, values := filter.buildWhereClause()

	query := `SELECT id, name, created_at, updated_at
              FROM indices %s`

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		values = append(values, limit, offset)
	}

	rows, err := ctx.SQL.QueryContext(ctx, fmt.Sprintf(query, whereClause), values...)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer rows.Close()

	var indices []*Index

	for rows.Next() {
		var i Index

		err = rows.Scan(&i.ID, &i.Name, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return nil, datasource.ErrorDB{Err: err}
		}

		indices = append(indices, &i)
	}

	if err = rows.Err(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	for _, index := range indices {
		index.Constituents, err = s.indexConstituentStore.List(ctx, &IndexConstituentFilter{IndexID: index.ID}, 0, 0)
		if err != nil {
			return nil, err
		}
	}

	return indices, nil
}

func (s *indexStore) Retrieve(ctx *gofr.Context, id int) (*Index, error) {
	var index Index

	query := `SELECT id, name, created_at, updated_at
              FROM indices WHERE id = ?`

	err := ctx.SQL.QueryRowContext(ctx, query, id).Scan(&index.ID, &index.Name, &index.CreatedAt, &index.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, http.ErrorEntityNotFound{Name: "indices", Value: strconv.Itoa(id)}
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	index.Constituents, err = s.indexConstituentStore.List(ctx, &IndexConstituentFilter{IndexID: id}, 0, 0)
	if err != nil {
		return nil, err
	}

	return &index, nil
}

func (s *indexStore) Create(ctx *gofr.Context, index *Index) (*Index, error) {
	txn, err := ctx.SQL.Begin()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer s.rollback(ctx, txn)

	query := "INSERT INTO indices (name, created_at, updated_at) VALUES (?, ?, ?)"

	result, err := txn.ExecContext(ctx, query, index.Name, index.CreatedAt, index.UpdatedAt)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	for i := range index.Constituents {
		index.Constituents[i].IndexID = int(id)
	}

	err = s.indexConstituentStore.CreateAll(ctx, txn, index.Constituents)
	if err != nil {
		return nil, err
	}

	if err = txn.Commit(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, int(id))
}

func (s *indexStore) Update(ctx *gofr.Context, id int, index *Index) (*Index, error) {
	txn, err := ctx.SQL.Begin()
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	defer s.rollback(ctx, txn)

	query := `UPDATE indices SET name = ?, created_at = ?, updated_at = ?
              WHERE id = ?`

	_, err = txn.ExecContext(ctx, query, index.Name, index.CreatedAt, index.UpdatedAt, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	err = s.indexConstituentStore.Delete(ctx, txn, id)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	err = s.indexConstituentStore.CreateAll(ctx, txn, index.Constituents)
	if err != nil {
		return nil, err
	}

	if err = txn.Commit(); err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	return s.Retrieve(ctx, id)
}

func (*indexStore) Delete(ctx *gofr.Context, id int) error {
	_, err := ctx.SQL.ExecContext(ctx, "DELETE FROM indices WHERE id = ?", id)
	if err != nil {
		return datasource.ErrorDB{Err: err}
	}

	return nil
}

func (*indexStore) rollback(ctx *gofr.Context, txn *gofrSQL.Tx) {
	if err := txn.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		ctx.Logger.Errorf("rollback failed: %v", err)
	}
}

func (f *IndexFilter) buildWhereClause() (clause string, values []any) {
	if f.Name != "" {
		clause += " AND name = ?"

		values = append(values, f.Name)
	}

	if clause != "" {
		clause = "WHERE" + strings.TrimPrefix(clause, " AND")
	}

	return clause, values
}
