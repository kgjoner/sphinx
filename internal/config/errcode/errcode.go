package errcode

const (
	InvalidCredentials = "INVALID_CREDENTIALS"
	InvalidAccess = "INVALID_ACCESS"
	ExpiredAccess = "EXPIRED_ACCESS"
	ExpiredSession = "EXPIRED_SESSION"

	AccountNotFound = "ACCOUNT_NOT_FOUND"
	ApplicationNotFound = "APPLICATION_NOT_FOUND"
	SessionNotFound = "SESSION_NOT_FOUND"

	DuplicateKey = "DUPLICATE_KEY"
	InvalidGranting = "INVALID_GRANTING"
)