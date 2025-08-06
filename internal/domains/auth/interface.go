package auth

import (
	"github.com/google/uuid"
)

type Repo interface {
	InsertUser(*User) error
	UpdateUser(User) error
	GetUserByID(uuid.UUID) (*User, error)
	GetUserByEntry(Entry) (*User, error)
	GetUserByLink(uuid.UUID) (*User, error)
	GetUserByExternalAuthID(provider string, id string) (*User, error)

	InsertApplication(*Application) error
	UpdateApplication(Application) error
	GetApplicationByID(uuid.UUID) (*Application, error)

	UpsertLinks(...Link) error
	UpsertSessions(...Session) error
}
