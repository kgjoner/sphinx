package identity

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/prim"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Repo interface {
	InsertUser(*User) error
	UpdateUser(User) error
	GetUserByID(uuid.UUID) (*User, error)
	GetUserByEntry(entry shared.Entry) (*User, error)
	GetUserByExternalCredential(providerName string, subjectID string) (*User, error)
	ListUsers(filter string, pag *prim.Pagination) (*prim.PaginatedData[User], error)

	InsertExternalCredential(*ExternalCredential) error
	UpdateExternalCredential(ExternalCredential) error
	RemoveExternalCredential(userID uuid.UUID, providerName string, subjectID string) error
}
