package types

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
)

type Actor struct {
	UserID        uuid.UUID
	Email         htypes.Email
	Name          string
	ApplicationID uuid.UUID
	SessionID     uuid.UUID
	Permissions   []string
}
