package shared

import (
	"context"
	"database/sql"
)

type RepoPool[T any] interface {
	NewDAO(context.Context) T
	WithTransaction(context.Context, *sql.TxOptions, func(T) (any, error)) (any, error)
	WithReadOnlyTransaction(context.Context, func(T) (any, error)) (any, error)
	Close() error
}