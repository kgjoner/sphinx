package baserepo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kgjoner/cornucopia/helpers/htypes"
	_ "github.com/lib/pq"
)

type Pool struct {
	url       string
	db        *sql.DB
}

// Creates a new pool for connections.
func NewPool(url string) (*Pool, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("baserepo: unable to open db: %v", err)
	}

	return &Pool{
		url: url,
		db:  db,
	}, nil
}

func (p Pool) DatabaseUrl() string {
	return p.url
}

type Queries struct {
	ctx   context.Context
	db    *sql.DB
}

func (p Pool) NewQueries(ctx context.Context) *Queries {
	return &Queries{
		ctx,
		p.db,
	}
}

func handleListQuery[T any](rows *sql.Rows, pag *htypes.Pagination, dest func(item *T) []any) (*htypes.PaginatedData[T], error) {
	items := []T{}
	for rows.Next() {
		var item T
		err := rows.Scan(dest(&item)...)
		if err == sql.ErrNoRows {
			return nil, nil
		} else if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(items) > pag.Limit {
		pag.HasNext = true
		items = items[:pag.Limit]
	}

	return htypes.NewPaginatedData(*pag, items), nil
}

func handleListQueryWithoutPagination[T any](rows *sql.Rows, dest func(item *T) []any) ([]T, error) {
	items := []T{}
	for rows.Next() {
		var item T
		err := rows.Scan(dest(&item)...)
		if err == sql.ErrNoRows {
			return nil, nil
		} else if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
