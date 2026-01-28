package shared

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
)

type Actor struct {
	ID          uuid.UUID
	Kind        SubjectKind
	Email       htypes.Email
	Name        string
	AudienceID  uuid.UUID
	SessionID   uuid.UUID
	Permissions []string
}

/* ==============================================================================
	Subject Kind
============================================================================== */

type SubjectKind string

const (
	KindUser SubjectKind = "user"
	KindApp  SubjectKind = "application"
)

func (i SubjectKind) Enumerate() any {
	return []SubjectKind{
		KindUser,
		KindApp,
	}
}
