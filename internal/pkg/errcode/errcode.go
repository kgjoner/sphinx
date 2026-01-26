package errcode

import "github.com/kgjoner/cornucopia/v2/helpers/apperr"

const (
	PermissionDenied   apperr.Code = "PERMISSION_DENIED"
	InvalidCredentials apperr.Code = "INVALID_CREDENTIALS"
	InvalidAccess      apperr.Code = "INVALID_ACCESS"
	ExpiredAccess      apperr.Code = "EXPIRED_ACCESS"
	ExpiredSession     apperr.Code = "EXPIRED_SESSION"

	DeactivatedUser   apperr.Code = "DEACTIVATED_USER"
	NoConsent         apperr.Code = "NO_CONSENT"
	RevokedConsent    apperr.Code = "REVOKED_CONSENT"
	NoRelatedProvider apperr.Code = "NO_RELATED_PROVIDER"

	InvalidProvider     apperr.Code = "INVALID_PROVIDER"
	UserNotFound        apperr.Code = "USER_NOT_FOUND"
	ApplicationNotFound apperr.Code = "APPLICATION_NOT_FOUND"
	LinkNotFound        apperr.Code = "LINK_NOT_FOUND"
	SessionNotFound     apperr.Code = "SESSION_NOT_FOUND"

	DuplicateKey apperr.Code = "DUPLICATE_KEY"
	InvalidRole  apperr.Code = "INVALID_ROLE"
)
