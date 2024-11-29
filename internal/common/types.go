package common

import (
	"context"

	"github.com/kgjoner/hermes/pkg/hermes"
	baserepo "github.com/kgjoner/sphinx/internal/repositories/base"
)

type RepoFactory[T any] interface {
	NewQueries(context.Context) T
}

type Pools struct {
	BasePool  RepoFactory[*baserepo.Queries]
}

type Services struct {
	MailService hermes.MailService
}
