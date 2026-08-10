package identity

import "github.com/kgjoner/cornucopia/v3/apperr"

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

	ErrEmailAlreadyVerified = apperr.NewRequestError(
		"email has already been verified",
		"identity.email_already_verified",
	)
	ErrPhoneAlreadyVerified = apperr.NewRequestError(
		"phone number has already been verified",
		"identity.phone_already_verified",
	)
	ErrNoPendingField = apperr.NewRequestError(
		"user does not have a pending field to cancel",
		"identity.no_pending_field",
	)
	ErrNoVerificationCode = apperr.NewConflictError(
		"user does not have a verification code.",
		"identity.no_verification_code",
	)
	ErrInvalidVerificationCode = apperr.NewRequestError(
		"verification code is invalid.",
		"identity.invalid_verification_code",
	)
)
