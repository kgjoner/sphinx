package sphinx

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/cornucopia/v2/utils/httputil"
)

// Get token owner's data.
func (s Service) Me(token string) (*User, error) {
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

// Get target user's data. Return error if target user does not exist.
//
// Token owner must be an admin.
func (s Service) User(userID uuid.UUID, token string) (*User, error) {
	var respData presenter.Success[User]
	_, err := s.httpApi.Get("/user/"+userID.String(), &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}

// Get target user's email. Return error if target user does not exist.
func (s Service) EmailOf(userID uuid.UUID) (htypes.Email, error) {
	var respData presenter.Success[htypes.Email]
	_, err := s.httpApi.Get("/user/"+userID.String()+"/email", &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
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
	_, err = s.httpApi.Post("/user", body, nil)(&respData)

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

// Get user id by their entry. Return zero value if entry is not found.
func (s Service) UserIDByEntry(entry string) (uuid.UUID, error) {
	var respData presenter.Success[uuid.UUID]
	_, err := s.httpApi.Get("/user/id", &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
			"X-Entry":       entry,
		},
	})(&respData)

	if err != nil {
		return uuid.Nil, err
	}

	return respData.Data, nil
}

// Add roles and/or grantings to target user.
func (s Service) GrantPermissions(userID uuid.UUID, roles []string) (bool, error) {
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
	_, err := s.httpApi.Patch("/user/"+userID.String()+"/permission", body, &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
		},
	})(&respData)

	if err != nil {
		return false, err
	}

	return respData.Data, nil
}

// Remove roles and/or grantings from target user.
func (s Service) RevokePermissions(userID uuid.UUID, roles []string) (bool, error) {
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
	_, err := s.httpApi.Patch("/user/"+userID.String()+"/permission", body, &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
		},
	})(&respData)

	if err != nil {
		return false, err
	}

	return respData.Data, nil
}
