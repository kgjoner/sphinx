package common

import (
	"context"

	"github.com/kgjoner/hermes/pkg/hermes"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type RepoFactory[T any] interface {
	New(context.Context) T
}

type RepoFactories struct {
	AuthRepo RepoFactory[authcase.AuthRepo]
}

type Services struct {
	MailService hermes.MailService
}
