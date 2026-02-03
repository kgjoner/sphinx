package auth

import "github.com/kgjoner/cornucopia/v2/helpers/apperr"

var (
	// Session related Errors
	ErrDeactivatedPrincipal = apperr.NewUnauthorizedError(
		"deactivated principal",
		"auth.deactivated_principal",
	)
	ErrInvalidAccess = apperr.NewUnauthorizedError(
		"invalid access",
		"auth.invalid_access",
	)
	ErrExpiredAccess = apperr.NewUnauthorizedError(
		"expired access",
		"auth.expired_access",
	)
	ErrInvalidSession = apperr.NewUnauthorizedError(
		"invalid session",
		"auth.invalid_session",
	)
	ErrExpiredSession = apperr.NewUnauthorizedError(
		"expired session",
		"auth.expired_session",
	)
	ErrSessionNotFound = apperr.NewUnauthorizedError(
		"session not found",
		"auth.session_not_found",
	)

	// OAuth2 Errors
	ErrNoConsent = apperr.NewForbiddenError(
		"user has not consented to this application",
		"auth.no_consent",
	)
	ErrInvalidClient = apperr.NewUnauthorizedError(
		"invalid client",
		"auth.invalid_client",
	)
	ErrInvalidRedirectURI = apperr.NewRequestError(
		"invalid redirect URI",
		"auth.invalid_redirect_uri",
	)
	ErrUnsupportedGrantType = apperr.NewRequestError(
		"unsupported grant type",
		"auth.unsupported_grant_type",
	)
	ErrExpiredGrant = apperr.NewUnauthorizedError(
		"expired grant",
		"auth.expired_grant",
	)
	ErrInvalidSolver = apperr.NewInternalError(
		"the provided solver is invalid; this is likely a programming error",
		"auth.invalid_solver",
	)

	// External Authentication Errors
	ErrNoUserForThisSubject = apperr.NewUnauthorizedError(
		"there is no user associated with this external subject",
		"auth.no_user_for_this_subject",
	)

	// Signing Key Errors
	ErrNoActiveKeys = apperr.NewInternalError(
		"no active signing key available",
		"auth.no_active_keys",
	)
	ErrRedundantRotation = apperr.NewRequestError(
		"signing key has already been rotated",
		"auth.redundant_rotation",
	)
	ErrKeyAlreadyExpired = apperr.NewRequestError(
		"signing key has already expired",
		"auth.key_already_expired",
	)
)
