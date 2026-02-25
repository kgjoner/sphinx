package access

import "github.com/kgjoner/cornucopia/v2/helpers/apperr"

var (
	ErrInvalidAppSecret = apperr.NewUnauthorizedError(
		"invalid application secret provided",
		"access.invalid_app_secret",
	)
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
