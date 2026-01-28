package access

import (
	"github.com/google/uuid"
)

type Repo interface {
	InsertApplication(*Application) error
	UpdateApplication(Application) error
	GetApplicationByID(uuid.UUID) (*Application, error)

	GetUserLink(userID uuid.UUID, appID uuid.UUID) (*Link, error)
	UpsertLinks(...Link) error
}
