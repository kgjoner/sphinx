package authcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type AuthRepo interface {
	InsertUser(*auth.User) error
	UpdateUser(auth.User) error
	GetUserByID(uuid.UUID) (*auth.User, error)
	GetUserByEntry(auth.Entry) (*auth.User, error)
	GetUserByLink(uuid.UUID) (*auth.User, error)

	InsertApplication(*auth.Application) error
	UpdateApplication(auth.Application) error
	GetApplicationByID(uuid.UUID) (*auth.Application, error)

	UpsertLinks(...auth.Link) error
	UpsertSessions(...auth.Session) error
}
