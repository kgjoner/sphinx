package auth

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/shared"
)

func CanIssueGrant(actor shared.Actor, subID uuid.UUID, audID uuid.UUID) error {
	if actor.Kind != shared.KindUser || actor.ID != subID || actor.AudienceID != audID {
		return shared.ErrNoPermission
	}

	return nil
}
