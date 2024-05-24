package authcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type AuthRepo interface {
	InsertAccount(auth.Account) (int, error)
	UpdateAccount(auth.Account) error
	GetAccountById(uuid.UUID) (*auth.Account, error)
	GetAccountByEntry(string) (*auth.Account, error)
	GetAccountByOAuthCode(string) (*auth.Account, error)

	InsertApplication(auth.Application) (int, error)
	UpdateApplication(auth.Application) error
	GetApplicationById(uuid.UUID) (*auth.Application, error)

	UpsertLinks(...auth.Link) error
	UpsertSessions(...auth.Session) error
}
