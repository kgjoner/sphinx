package access

import "github.com/kgjoner/cornucopia/v3/apperr"

var (
	ErrApplicationNotFound = apperr.NewRequestError(
		"application not found",
		"access.application_not_found",
	)
	ErrLinkNotFound = apperr.NewRequestError(
		"link not found",
		"access.link_not_found",
	)
	ErrEmptyRole = apperr.NewValidationError(
		"role must be provided",
		"access.empty_role",
	)
	ErrInvalidRole = apperr.NewRequestError(
		"application does not support the desired role",
		"access.invalid_role",
	)
	ErrRedundantRequest = apperr.NewRequestError(
		"redundant request",
		"access.redundant_request",
	)
)
