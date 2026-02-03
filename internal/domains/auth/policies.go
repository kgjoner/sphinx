package auth

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/shared"
)

const (
	PermKeysManage     = "auth.keys:manage"
	PermKeysReadStatus = "auth.keys:read_status"
	PermKeysReadAll    = "auth.keys:read_all"
)

func IsRootApp(appID uuid.UUID) bool {
	return appID.String() == config.Env.ROOT_APP_ID
}

func CanIssueGrant(actor shared.Actor, targetSubID uuid.UUID, targetAudID uuid.UUID) error {
	// Only users with root application tokens can issue grants to themselves for non-root applications
	hasPermission := actor.Kind == shared.KindUser && actor.ID == targetSubID &&
		IsRootApp(actor.AudienceID) && !IsRootApp(targetAudID)

	if !hasPermission {
		return shared.ErrNoPermission
	}

	return nil
}

func CanManageKeys(actor shared.Actor) error {
	if !IsRootApp(actor.AudienceID) {
		return shared.ErrNoPermission
	}

	var hasPermission bool
	for _, p := range actor.Permissions {
		if p == PermKeysManage {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return shared.ErrNoPermission
	}

	return nil
}

func CanReadKeyStatus(actor shared.Actor) error {
	if !IsRootApp(actor.AudienceID) {
		return shared.ErrNoPermission
	}

	var hasPermission bool
	for _, p := range actor.Permissions {
		if p == PermKeysReadStatus || p == PermKeysReadAll {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return shared.ErrNoPermission
	}

	return nil
}

func CanReadAllKeys(actor shared.Actor) error {
	if !IsRootApp(actor.AudienceID) {
		return shared.ErrNoPermission
	}

	var hasPermission bool
	for _, p := range actor.Permissions {
		if p == PermKeysReadAll {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return shared.ErrNoPermission
	}

	return nil
}
