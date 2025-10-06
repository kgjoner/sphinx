package sphinx

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/utils/httputil"
)

// Get token owner's data.
func (s Service) User(token string) (*User, error) {
	var respData presenter.Success[User]
	_, err := s.httpApi.Get("/user", &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}

// Get target user's data. Target value can be any user entry, including ID. Return error if target user does not exist.
//
// Token owner must be an admin.
func (s Service) UserOf(target string, token string) (*User, error) {
	var respData presenter.Success[User]
	_, err := s.httpApi.Get("/user", &httputil.Options{
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

// Get target user's email. Target value can be their ID or other entry. Return error if target user does not exist.
func (s Service) EmailOf(target string) (htypes.Email, error) {
	var respData presenter.Success[htypes.Email]
	_, err := s.httpApi.Get("/user/email", &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
			"X-Target":      target,
		},
	})(&respData)

	if err != nil {
		return "", err
	}

	return respData.Data, nil
}

// Create a simple user for the informed email.
func (s Service) NewUser(email htypes.Email, password string) (userID uuid.UUID, err error) {
	body := map[string]any{
		"email":    email,
		"password": password,
	}

	var respData presenter.Success[User]
	_, err = s.httpApi.Post("/user", body, &httputil.Options{
		Headers: map[string]string{
			"X-App": s.appID,
		},
	})(&respData)

	if err != nil {
		return uuid.Nil, err
	}

	return respData.Data.ID, nil
}

// Check whether entry exists.
func (s Service) DoesEntryExist(entry string) (bool, error) {
	var respData presenter.Success[bool]
	_, err := s.httpApi.Get("/user/existence", &httputil.Options{
		Headers: map[string]string{
			"X-Entry": entry,
		},
	})(&respData)

	if err != nil {
		return false, err
	}

	return respData.Data, nil
}

// Get user id by their entry. Return nil if entry is not found.
func (s Service) UserIDByEntry(entry string) (*uuid.UUID, error) {
	var respData presenter.Success[*uuid.UUID]
	_, err := s.httpApi.Get("/user/id", &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
			"X-Entry":       entry,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	return respData.Data, nil
}

// Add roles and/or grantings to target user.
func (s Service) GrantPermissions(target string, roles []string) (bool, error) {
	body := map[string]any{
		"shouldRemove": sql.NullBool{
			Valid: true,
			Bool:  false,
		},
	}

	if len(roles) > 0 {
		body["roles"] = roles
	}

	var respData presenter.Success[bool]
	_, err := s.httpApi.Patch("/user/permission", body, &httputil.Options{
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

// Remove roles and/or grantings from target user.
func (s Service) RevokePermissions(target string, roles []string) (bool, error) {
	body := map[string]any{
		"shouldRemove": sql.NullBool{
			Valid: true,
			Bool:  true,
		},
	}

	if len(roles) > 0 {
		body["roles"] = roles
	}

	var respData presenter.Success[bool]
	_, err := s.httpApi.Patch("/user/permission", body, &httputil.Options{
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
