package common

import (
	"context"
	"database/sql"

	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/hermes/pkg/hermes"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type RepoPool[T any] interface {
	NewQueries(context.Context) T
	WithTransaction(context.Context, *sql.TxOptions, func(T) (any, error)) (any, error)
	WithReadOnlyTransaction(context.Context, func(T) (any, error)) (any, error)
	Close() error
}

type BaseRepo interface {
	authcase.AuthRepo
}

type Pools struct {
	BasePool  RepoPool[BaseRepo]
	CachePool cache.Pool
}

type Services struct {
	MailService hermes.MailService
}
