package auth

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/shared"
)

func CanIssueGrant(actor shared.Actor, targetSubID uuid.UUID, targetAudID uuid.UUID) error {
	// Only users with root application tokens can issue grants to themselves for non-root applications
	if actor.Kind != shared.KindUser || actor.ID != targetSubID ||
		actor.AudienceID.String() != config.Env.ROOT_APP_ID || targetAudID.String() == config.Env.ROOT_APP_ID {
		return shared.ErrNoPermission
	}

	return nil
}
