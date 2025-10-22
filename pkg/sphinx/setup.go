package sphinx

import (
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/cornucopia/v2/utils/httputil"
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

	svc := &Service{
		httpApi:   httpApi,
		appID:     appID,
		appSecret: appSecret,
		appToken:  appToken,
	}

	return svc
}

type User struct {
	ID       uuid.UUID          `json:"id" validate:"required"`
	Email    htypes.Email       `json:"email" validate:"required"`
	Phone    htypes.PhoneNumber `json:"phone,omitempty"`
	Username string             `json:"username,omitempty" validate:"wordID"`
	Document htypes.Document    `json:"document,omitempty"`
	Name     string             `json:"name,omitempty"`
	Surname  string             `json:"surname,omitempty"`
	Address  htypes.Address     `json:"address,omitempty"`

	IsActive             bool           `json:"isActive"`
	HasEmailBeenVerified bool           `json:"hasEmailBeenVerified"`
	HasPhoneBeenVerified bool           `json:"hasPhoneBeenVerified"`
	Link                 *auth.LinkView `json:"link"`
}

func (a User) DisplayName() string {
	if a.Name != "" {
		return a.Name
	}

	if a.Username != "" {
		return a.Username
	}

	email := a.Email.String()
	if at := strings.IndexByte(email, '@'); at > 0 {
		return email[:at]
	}
	return email
}

func (a User) IsAdmin() bool {
	for _, r := range a.Link.Roles {
		if r == auth.RoleAdmin {
			return true
		}
	}

	return false
}

func (a User) IsDev() bool {
	for _, r := range a.Link.Roles {
		if r == auth.RoleDev {
			return true
		}
	}

	return false
}

func (a User) HasRole(role string) bool {
	for _, r := range a.Link.Roles {
		if string(r) == role {
			return true
		}
	}

	return false
}
