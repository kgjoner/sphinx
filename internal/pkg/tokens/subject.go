package tokens

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
)

type Subject struct {
	SessionID     uuid.UUID
	UserID        uuid.UUID
	UserEmail     htypes.Email
	UserName      string
	ApplicationID uuid.UUID
	Roles         []string
}
