package pgpool

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Pool struct {
	url string
	db  *sql.DB
}

// Creates a new pool for connections.
func New(url string) (*Pool, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("pgpool: unable to open db: %v", err)
	}

	return &Pool{
		url: url,
		db:  db,
	}, nil
}

func (p *Pool) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *Pool) DatabaseURL() string {
	return p.url
}

func (p *Pool) Connection() *sql.DB {
	return p.db
}

// Creates a new SQL transaction enabled to be used in the function passed.
// If opts is nil, it defaults to READ COMMITTED isolation level.
//
// If an error occurs, the transaction is rolled back.
// If successful, the transaction is committed.
// It returns the result of the function or an error if it fails.
func (p *Pool) WithTx(ctx context.Context, opts *sql.TxOptions, fn func(*sql.Tx) (any, error)) (any, error) {
	if opts == nil {
		opts = &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		}
	}

	tx, err := p.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("pgpool: unable to begin transaction: %v", err)
	}

	output, err := fn(tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, fmt.Errorf("pgpool: transaction rollback failed: %v, original error: %v", rollbackErr, err)
		}
		return nil, err
	}

	return output, tx.Commit()
}

// Creates a new SQL read-only transaction. It has REPEATABLE READ isolation level.
// This is useful for operations that require consistent reads without modifying data.
func (p *Pool) WithReadOnlyTx(ctx context.Context, fn func(*sql.Tx) (any, error)) (any, error) {
	opts := &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	}

	return p.WithTx(ctx, opts, fn)
}
