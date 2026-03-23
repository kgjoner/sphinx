package shared

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/prim"
)

type Actor struct {
	ID          uuid.UUID
	Kind        SubjectKind
	Email       prim.Email
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
	KindUser   SubjectKind = "user"
	KindApp    SubjectKind = "application"
	KindSystem SubjectKind = "system"
)

func (i SubjectKind) Enumerate() any {
	return []SubjectKind{
		KindUser,
		KindApp,
		KindSystem,
	}
}
