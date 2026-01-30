package sphinx

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/cornucopia/v2/utils/httputil"
)

type ExternalCredentialView struct {
	UserID            uuid.UUID `json:"userId"`
	ProviderName      string    `json:"providerName"`
	ProviderSubjectID string    `json:"providerSubjectId"`
	ProviderAlias     string    `json:"providerAlias"`
	LastUsedAt        time.Time `json:"lastUsedAt"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type UserView struct {
	ID       uuid.UUID          `json:"id"`
	Email    htypes.Email       `json:"email"`
	Phone    htypes.PhoneNumber `json:"phone,omitempty"`
	Username string             `json:"username,omitempty"`
	Document htypes.Document    `json:"document,omitempty"`
	Name     string             `json:"name,omitempty"`
	Surname  string             `json:"surname,omitempty"`
	Address  *htypes.Address    `json:"address,omitempty"`

	PendingEmail         htypes.Email       `json:"pendingEmail,omitempty"`
	HasEmailBeenVerified bool               `json:"hasEmailBeenVerified"`
	PendingPhone         htypes.PhoneNumber `json:"pendingPhone,omitempty"`
	HasPhoneBeenVerified bool               `json:"hasPhoneBeenVerified"`
	UsernameUpdatedAt    htypes.NullTime    `json:"usernameUpdatedAt"`

	ExternalCredentials []ExternalCredentialView `json:"externalCredentials,omitempty"`
	IsActive            bool                     `json:"isActive"`
}

type UserLeanView struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username,omitempty"`
	Name     string    `json:"name,omitempty"`
	Surname  string    `json:"surname,omitempty"`

	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Get target user's data. Return error if target user does not exist.
func (s Client) User(userID uuid.UUID) (*UserView, error) {
	var respData presenter.Success[UserView]
	_, err := s.httpApi.Get("/user/"+userID.String(), &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}

// Get target user's email. Return error if target user does not exist.
func (s Client) EmailOf(userID uuid.UUID) (htypes.Email, error) {
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
func (s Client) NewUser(email htypes.Email, password string) (userID uuid.UUID, err error) {
	body := map[string]any{
		"email":    email,
		"password": password,
	}

	var respData presenter.Success[UserView]
	_, err = s.httpApi.Post("/user", body, nil)(&respData)

	if err != nil {
		return uuid.Nil, err
	}

	return respData.Data.ID, nil
}

// Check whether entry exists.
func (s Client) DoesEntryExist(entry string) (bool, error) {
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
func (s Client) UserIDByEntry(entry string) (uuid.UUID, error) {
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

// Add role to an user.
func (s Client) AddRole(userID uuid.UUID, role string) error {
	_, err := s.httpApi.Put("/user/"+userID.String()+"/link/"+s.appID+"/role/"+role, nil, &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
		},
	})(nil)

	return err
}

// Remove role from an user.
func (s Client) RemoveRole(userID uuid.UUID, role string) error {
	_, err := s.httpApi.Delete("/user/"+userID.String()+"/link/"+s.appID+"/role/"+role, &httputil.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
		},
	})(nil)

	return err
}
