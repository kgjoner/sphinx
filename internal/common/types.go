package common

import (
	"context"

	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/hermes/pkg/hermes"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type RepoFactory[T any] interface {
	NewQueries(context.Context) T
}

type BaseRepo interface {
	authcase.AuthRepo
}

type Pools struct {
	BasePool  RepoFactory[BaseRepo]
	CachePool cache.Pool
}

type Services struct {
	MailService hermes.MailService
}
