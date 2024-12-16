package sphinx

import (
	"database/sql"

	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/utils/httputil"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

// Get token owner's data.
func (s SphinxService) Account(token string) (*auth.AccountPrivateView, error) {
	var respData presenter.Success[auth.AccountPrivateView]
	_, err := s.httpApi.Get("/account", &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}

// Get target account's data. Target value can be any account entry, including ID. Token owner must be an admin.
func (s SphinxService) AccountOf(target string, token string) (*auth.AccountPrivateView, error) {
	var respData presenter.Success[auth.AccountPrivateView]
	_, err := s.httpApi.Get("/account", &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
			"X-Target":      target,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}

// Add roles and/or grantings to target account.
func (s SphinxService) GrantPermissions(target string, roles []string, grantings []string) (bool, error) {
	body := map[string]any{
		"shouldRemove": sql.NullBool{
			Valid: true,
			Bool:  false,
		},
	}

	if len(roles) > 0 {
		body["roles"] = roles
	}

	if len(grantings) > 0 {
		body["grantings"] = grantings
	}

	var respData presenter.Success[bool]
	_, err := s.httpApi.Patch("/account/permission", body, &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
			"X-Target":      target,
		},
	})(&respData)

	if err != nil {
		return false, err
	}

	return respData.Data, nil
}

// Remove roles and/or grantings from target account.
func (s SphinxService) RevokePermissions(target string, roles []string, grantings []string) (bool, error) {
	body := map[string]any{
		"shouldRemove": sql.NullBool{
			Valid: true,
			Bool:  true,
		},
	}

	if len(roles) > 0 {
		body["roles"] = roles
	}

	if len(grantings) > 0 {
		body["grantings"] = grantings
	}

	var respData presenter.Success[bool]
	_, err := s.httpApi.Patch("/account/permission", body, &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
			"X-Target":      target,
		},
	})(&respData)

	if err != nil {
		return false, err
	}

	return respData.Data, nil
}
