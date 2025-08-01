package sphinx

import (
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/utils/httputil"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Service struct {
	httpApi   *httputil.HTTPUtil
	appID     string
	appSecret string
	appToken  string
}

func New(baseURL, appID, appSecret string) *Service {
	httpApi := httputil.New(baseURL)
	appToken := base64.StdEncoding.EncodeToString([]byte(appID + ":" + appSecret))

	return &Service{
		httpApi:   httpApi,
		appID:     appID,
		appSecret: appSecret,
		appToken:  appToken,
	}
}

type Account struct {
	ID       uuid.UUID          `json:"id" validate:"required"`
	Email    htypes.Email       `json:"email" validate:"required"`
	Phone    htypes.PhoneNumber `json:"phone,omitempty"`
	Username string             `json:"username,omitempty" validate:"wordID"`
	Document htypes.Document    `json:"document,omitempty"`
	Name     string             `json:"name,omitempty"`
	Surname  string             `json:"surname,omitempty"`
	Address  htypes.Address     `json:"address,omitempty"`

	IsActive             bool       `json:"isActive"`
	HasEmailBeenVerified bool       `json:"hasEmailBeenVerified"`
	HasPhoneBeenVerified bool       `json:"hasPhoneBeenVerified"`
	Link                 *auth.Link `json:"link"`
}

func (a Account) IsAdmin() bool {
	for _, r := range a.Link.Roles {
		if r == auth.RoleAdmin {
			return true
		}
	}

	return false
}

func (a Account) IsDev() bool {
	for _, r := range a.Link.Roles {
		if r == auth.RoleDev {
			return true
		}
	}

	return false
}

func (a Account) HasRole(role string) bool {
	for _, r := range a.Link.Roles {
		if string(r) == role {
			return true
		}
	}

	return false
}
