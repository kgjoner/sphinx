package shared

import "github.com/kgjoner/cornucopia/v2/helpers/apperr"

var (
	ErrNoPermission = apperr.NewForbiddenError(
		"you do not have permission to perform this action",
		"no_permission",
	)
	ErrMissingCredentials = apperr.NewUnauthorizedError(
		"missing credentials",
		"missing_credentials",
	)
	ErrInvalidCredentials = apperr.NewUnauthorizedError(
		"invalid credentials",
		"invalid_credentials",
	)
	ErrEmptyPassword = apperr.NewValidationError(
		"password cannot be empty",
		"empty_password",
	)
	ErrEmptyHashedData = apperr.NewValidationError(
		"hash cannot be empty",
		"empty_hashed_data",
	)
	ErrInvalidCode = apperr.NewValidationError(
		"invalid or expired code; please request a new one",
		"invalid_code",
	)
	ErrInvalidProof = apperr.NewInternalError(
		"this action does not accept provided proof of authentication; this is likely a programming error",
		"invalid_proof",
	)
	ErrInvalidExternalSubject = apperr.NewInternalError(
		"external subject does not contain required information",
		"invalid_external_subject",
	)
)
