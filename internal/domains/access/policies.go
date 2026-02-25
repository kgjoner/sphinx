package access

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/shared"
)

const (
	PermAppCreate         = "access.application:create"
	PermAppEdit           = "access.application:edit"
	PermAppRecreateSecret = "access.application:recreate_secret"
	PermAppReadAll        = "access.application:read_all"
	PermLinkManage        = "access.link:manage"
	PermRoleManageExtra   = "access.role:manage_extra"
	PermRoleManageManager = "access.role:manage_manager"
	PermRoleManageAdmin   = "access.role:manage_admin"
	PermLinkReadAll       = "access.link:read_all"
)

func IsRootApp(appID uuid.UUID) bool {
	return appID.String() == config.Env.ROOT_APP_ID
}

/* ==============================================================================
	APPLICATION PERMISSION CHECKS
============================================================================== */

func CanCreateApplication(a *shared.Actor) error {
	if a.AudienceID == uuid.Nil || !IsRootApp(a.AudienceID) {
		return shared.ErrNoPermission
	}

	hasPermission := false
	for _, p := range a.Permissions {
		if p == PermAppCreate {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return shared.ErrNoPermission
	}

	return nil
}

func CanEditApplication(a *shared.Actor, targetID uuid.UUID) error {
	if a.AudienceID == uuid.Nil {
		return shared.ErrNoPermission
	}

	if a.AudienceID != targetID {
		return shared.ErrNoPermission
	}

	hasPermission := false
	for _, p := range a.Permissions {
		if p == PermAppEdit {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return shared.ErrNoPermission
	}

	return nil
}

func CanRecreateAppSecret(a *shared.Actor, targetID uuid.UUID) error {
	if a.AudienceID == uuid.Nil {
		return shared.ErrNoPermission
	}

	if a.AudienceID != targetID {
		return shared.ErrNoPermission
	}

	hasPermission := false
	for _, p := range a.Permissions {
		if p == PermAppRecreateSecret {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return shared.ErrNoPermission
	}

	return nil
}

func CanReadApplications(a *shared.Actor) error {
	hasPermission := false
	for _, p := range a.Permissions {
		if p == PermAppReadAll {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return shared.ErrNoPermission
	}

	return nil
}

/* ==============================================================================
	LINK PERMISSION CHECKS
============================================================================== */

func CanCreateLink(a *shared.Actor, targetUserID uuid.UUID, targetAppID uuid.UUID) error {
	if a.AudienceID == uuid.Nil {
		return shared.ErrNoPermission
	}

	// If the actor is the same user, they can create a link only if the application is root.
	// (They don't have same application access before link creation.)
	if a.Kind == shared.KindUser && a.ID == targetUserID {
		if !IsRootApp(a.AudienceID) {
			return shared.ErrNoPermission
		}
		
		return nil
	}

	// For other actors, they need the permission and the application must match.
	if a.AudienceID != targetAppID {
		return shared.ErrNoPermission
	}

	hasPermission := false
	for _, p := range a.Permissions {
		if p == PermLinkManage {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return shared.ErrNoPermission
	}

	return nil
}

func CanManageRole(a *shared.Actor, targetAppID uuid.UUID, role Role) error {
	if a.AudienceID == uuid.Nil || a.AudienceID != targetAppID {
		return shared.ErrNoPermission
	}

	requiredPermission := PermRoleManageExtra
	switch role {
	case Manager:
		requiredPermission = PermRoleManageManager
	case Admin:
		requiredPermission = PermRoleManageAdmin
	}

	hasPermission := false
	for _, p := range a.Permissions {
		if p == requiredPermission {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return shared.ErrNoPermission
	}
	
	return nil
}

func CanManageConsent(a *shared.Actor, targetUserID uuid.UUID, targetAppID uuid.UUID) error {
	if a.AudienceID == uuid.Nil || a.AudienceID != targetAppID {
		return shared.ErrNoPermission
	}

	if a.Kind != shared.KindUser || a.ID != targetUserID {
		return shared.ErrNoPermission
	}
	
	return nil
}

func CanReadLink(a *shared.Actor, targetUserID uuid.UUID, targetAppID uuid.UUID) error {
	if  a.Kind == shared.KindUser && a.ID == targetUserID {
		return nil
	}

	if a.AudienceID == uuid.Nil || a.AudienceID != targetAppID {
		return shared.ErrNoPermission
	}

	hasPermission := false
	for _, p := range a.Permissions {
		if p == PermLinkReadAll {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return shared.ErrNoPermission
	}

	return nil
}
