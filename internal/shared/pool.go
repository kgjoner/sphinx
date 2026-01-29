package shared

import (
	"context"
	"database/sql"
)

type RepoPool interface {
	Connection() *sql.DB
	WithTx(context.Context, *sql.TxOptions, func(*sql.Tx) (any, error)) (any, error)
	WithReadOnlyTx(context.Context, func(*sql.Tx) (any, error)) (any, error)
	Close() error
}

type DBTX interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type RepoFactory[T any] interface {
	NewDAO(context.Context, DBTX) T
}
