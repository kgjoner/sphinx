package sphinx

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/utils/httputil"
)

// Get token owner's data.
func (s Service) Account(token string) (*Account, error) {
	var respData presenter.Success[Account]
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

// Get target account's data. Target value can be any account entry, including ID. Return error if target account does not exist.
//
// Token owner must be an admin.
func (s Service) AccountOf(target string, token string) (*Account, error) {
	var respData presenter.Success[Account]
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

// Get target account's email. Target value can be their ID or other entry. Return error if target account does not exist.
func (s Service) EmailOf(target string) (htypes.Email, error) {
	var respData presenter.Success[htypes.Email]
	_, err := s.httpApi.Get("/account/email", &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
			"X-Target": target,
		},
	})(&respData)

	if err != nil {
		return "", err
	}

	return respData.Data, nil
}

// Create a simple account for the informed email.
func (s Service) NewAccount(email htypes.Email, password string) (accountId uuid.UUID, err error) {
	body := map[string]any{
		"email": email,
		"password": password,
	}

	var respData presenter.Success[Account]
	_, err = s.httpApi.Post("/account", body, &httputil.Options{
		Headers: map[string]string{
			"X-App": s.appId,
		},
	})(&respData)

	if err != nil {
		return uuid.Nil, err
	}

	return respData.Data.Id, nil
}

// Check whether entry exists.
func (s Service) DoesEntryExist(entry string) (bool, error) {
	var respData presenter.Success[bool]
	_, err := s.httpApi.Get("/account/existence", &httputil.Options{
		Headers: map[string]string{
			"X-Entry": entry,
		},
	})(&respData)

	if err != nil {
		return false, err
	}

	return respData.Data, nil
}

// Get account id by their entry. Return nil if entry is not found.
func (s Service) AccountIdByEntry(entry string) (*uuid.UUID, error) {
	var respData presenter.Success[*uuid.UUID]
	_, err := s.httpApi.Get("/account/id", &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
			"X-Entry": entry,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	return respData.Data, nil
}

// Add roles and/or grantings to target account.
func (s Service) GrantPermissions(target string, roles []string, grantings []string) (bool, error) {
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
func (s Service) RevokePermissions(target string, roles []string, grantings []string) (bool, error) {
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
