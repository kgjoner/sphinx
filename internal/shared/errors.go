package shared

import "github.com/kgjoner/cornucopia/v2/helpers/apperr"

var (
	ErrNoPermission = apperr.NewForbiddenError(
		"you do not have permission to perform this action",
		"shared.no_permission",
	)
	ErrMissingCredentials = apperr.NewUnauthorizedError(
		"missing credentials",
		"shared.missing_credentials",
	)
	ErrInvalidCredentials = apperr.NewUnauthorizedError(
		"invalid credentials",
		"shared.invalid_credentials",
	)
	ErrEmptyPassword = apperr.NewValidationError(
		"password cannot be empty",
		"shared.empty_password",
	)
	ErrEmptyHashedData = apperr.NewValidationError(
		"hash cannot be empty",
		"shared.empty_hashed_data",
	)
	ErrInvalidCode = apperr.NewValidationError(
		"invalid or expired code; please request a new one",
		"shared.invalid_code",
	)
	ErrInvalidProof = apperr.NewInternalError(
		"this action does not accept provided proof of authentication; this is likely a programming error",
		"shared.invalid_proof",
	)
)
