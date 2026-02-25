package identity

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/shared"
)

const (
	PermUserWriteAll = "identity.user:write_all"
	PermUserReadAll  = "identity.user:read_all"
	PermUserList     = "identity.user:list"
)

func CanUpdateUser(actor *shared.Actor, targetID uuid.UUID) error {
	var hasPermission bool
	for _, p := range actor.Permissions {
		if p == PermUserWriteAll {
			hasPermission = true
			break
		}
	}

	if !hasPermission && (actor.Kind != shared.KindUser || actor.ID != targetID) {
		return shared.ErrNoPermission
	}

	return nil
}

func CanUpdateUsername(actor *shared.Actor, target *User) error {
	err := CanUpdateUser(actor, target.ID)
	if err != nil {
		return err
	}

	now := time.Now()
	if target.UsernameUpdatedAt.Time.After(now.Add(-time.Hour * 24 * 90)) {
		return ErrUsernameCooldown
	}

	return nil
}

func CanManageExternalCredentials(actor *shared.Actor, targetID uuid.UUID) error {
	if actor.Kind != shared.KindUser || actor.ID != targetID {
		return shared.ErrNoPermission
	}

	return nil
}

func CanReadUserSensitiveData(actor *shared.Actor, targetID uuid.UUID) error {
	var hasPermission bool
	for _, p := range actor.Permissions {
		if p == PermUserReadAll {
			hasPermission = true
			break
		}
	}

	if !hasPermission && (actor.Kind != shared.KindUser || actor.ID != targetID) {
		return shared.ErrNoPermission
	}

	return nil
}

func CanListUsers(actor *shared.Actor, filter string) error {
	if actor == nil {
		return shared.ErrNoPermission
	}

	// If the search filter is empty or too short, require special permission
	if len(filter) < 4 {
		hasPermission := false
		for _, p := range actor.Permissions {
			if p == PermUserList {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return shared.ErrNoPermission
		}
	}

	return nil
}
