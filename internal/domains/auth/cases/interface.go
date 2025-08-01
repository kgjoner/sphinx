package authcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type AuthRepo interface {
	InsertAccount(*auth.Account) error
	UpdateAccount(auth.Account) error
	GetAccountByID(uuid.UUID) (*auth.Account, error)
	GetAccountByEntry(auth.Entry) (*auth.Account, error)
	GetAccountByLink(uuid.UUID) (*auth.Account, error)

	InsertApplication(*auth.Application) error
	UpdateApplication(auth.Application) error
	GetApplicationByID(uuid.UUID) (*auth.Application, error)

	UpsertLinks(...auth.Link) error
	UpsertSessions(...auth.Session) error
}
