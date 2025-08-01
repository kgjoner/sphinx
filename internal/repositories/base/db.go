package baserepo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kgjoner/sphinx/internal/common"
	_ "github.com/lib/pq"
)

type Pool struct {
	url string
	db  *sql.DB
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

func (p Pool) DatabaseURL() string {
	return p.url
}

type DAO struct {
	ctx context.Context
	db  *sql.DB
	tx  *sql.Tx
}

func (p Pool) NewDAO(ctx context.Context) common.BaseRepo {
	return &DAO{
		ctx,
		p.db,
		nil,
	}
}

// Creates a new DAO with transaction enabled to be used in the function passed.
// If opts is nil, it defaults to READ COMMITTED isolation level.
//
// If an error occurs, the transaction is rolled back.
// If successful, the transaction is committed.
// The function fn receives a BaseRepo interface to perform operations within the transaction.
// It returns the result of the function or an error if it fails.
func (p Pool) WithTransaction(ctx context.Context, opts *sql.TxOptions, fn func(common.BaseRepo) (any, error)) (any, error) {
	if opts == nil {
		opts = &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		}
	}

	tx, err := p.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("baserepo: unable to begin transaction: %v", err)
	}

	dao := &DAO{
		ctx: ctx,
		db:  p.db,
		tx:  tx,
	}

	output, err := fn(dao)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, fmt.Errorf("baserepo: transaction rollback failed: %v, original error: %v", rollbackErr, err)
		}
		return nil, fmt.Errorf("baserepo: transaction failed: %v", err)
	}

	return output, tx.Commit()
}

// Creates a new DAO with a read-only transaction enabled. It has REPEATABLE READ isolation level.
// This is useful for operations that require consistent reads without modifying data.
func (p Pool) WithReadOnlyTransaction(ctx context.Context, fn func(common.BaseRepo) (any, error)) (any, error) {
	opts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	return p.WithTransaction(ctx, opts, fn)
}
