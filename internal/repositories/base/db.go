package baserepo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kgjoner/sphinx/internal/common"
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

func (p Pool) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p Pool) DatabaseUrl() string {
	return p.url
}

type Queries struct {
	ctx   context.Context
	db    *sql.DB
}

func (p Pool) NewQueries(ctx context.Context) common.BaseRepo {
	return &Queries{
		ctx,
		p.db,
	}
}
