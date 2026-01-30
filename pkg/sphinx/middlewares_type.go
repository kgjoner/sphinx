package sphinx

import (
	"net/http"
	"strings"

	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

var (
	ErrInvalidAccess = auth.ErrInvalidAccess
)

/* ==============================================================================
	Authorizer
============================================================================== */

type Authorizer interface {
	AuthorizeSubject(sub Subject, r *http.Request) (*http.Request, error)
}

/* ==============================================================================
	Subject
============================================================================== */

type Subject auth.Subject

func (a Subject) DisplayName() string {
	if a.Name != "" {
		return a.Name
	}

	email := a.Email.String()
	if at := strings.IndexByte(email, '@'); at > 0 {
		return email[:at]
	}
	return email
}

func (a Subject) IsAdmin() bool {
	for _, r := range a.Roles {
		if r == string(access.Admin) {
			return true
		}
	}

	return false
}

func (a Subject) IsManager() bool {
	for _, r := range a.Roles {
		if r == string(access.Manager) {
			return true
		}
	}

	return false
}

func (a Subject) HasRole(role string) bool {
	for _, r := range a.Roles {
		if r == role {
			return true
		}
	}

	return false
}
