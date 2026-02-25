package identity

import "github.com/kgjoner/cornucopia/v2/helpers/apperr"

var (
	ErrEmptyInput = apperr.NewValidationError(
		"input cannot be empty",
		"identity.empty_input",
	)
	ErrRedundantRequest = apperr.NewRequestError(
		"one or more fields are already set to the same value",
		"identity.redundant_request",
	)
	ErrInvalidField = apperr.NewRequestError(
		"target field does not exist or cannot be used on this action",
		"identity.invalid_field",
	)
	ErrUsernameCooldown = apperr.NewRequestError(
		"username can only be updated once every 90 days",
		"identity.username_cooldown",
	)
	ErrUserNotFound = apperr.NewRequestError(
		"user not found",
		"identity.user_not_found",
	)
	ErrDuplicateEntry = apperr.NewConflictError(
		"user already exists",
		"identity.duplicate_entry",
	)
	ErrExistingExternalCredential = apperr.NewConflictError(
		"external credential is already linked to another user",
		"identity.existing_external_credential",
	)
	ErrExternalCredentialNotFound = apperr.NewRequestError(
		"external credential not found",
		"identity.external_credential_not_found",
	)
)
