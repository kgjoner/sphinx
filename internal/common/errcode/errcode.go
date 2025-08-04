package errcode

const (
	InvalidCredentials = "INVALID_CREDENTIALS"
	InvalidAccess      = "INVALID_ACCESS"
	ExpiredAccess      = "EXPIRED_ACCESS"
	ExpiredSession     = "EXPIRED_SESSION"
	DeactivatedUser    = "DEACTIVATED_USER"
	NoConsent          = "NO_CONSENT"
	RevokedConsent     = "REVOKED_CONSENT"

	UserNotFound        = "USER_NOT_FOUND"
	ApplicationNotFound = "APPLICATION_NOT_FOUND"
	SessionNotFound     = "SESSION_NOT_FOUND"

	DuplicateKey = "DUPLICATE_KEY"
	InvalidRole  = "INVALID_ROLE"
)
