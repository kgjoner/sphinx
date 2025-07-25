package errcode

const (
	InvalidCredentials = "INVALID_CREDENTIALS"
	InvalidAccess      = "INVALID_ACCESS"
	ExpiredAccess      = "EXPIRED_ACCESS"
	ExpiredSession     = "EXPIRED_SESSION"
	DeactivatedAccount = "DEACTIVATED_ACCOUNT"
	NoConsent          = "NO_CONSENT"
	RevokedConsent     = "REVOKED_CONSENT"

	AccountNotFound     = "ACCOUNT_NOT_FOUND"
	ApplicationNotFound = "APPLICATION_NOT_FOUND"
	SessionNotFound     = "SESSION_NOT_FOUND"

	DuplicateKey = "DUPLICATE_KEY"
	InvalidRole  = "INVALID_ROLE"
)
