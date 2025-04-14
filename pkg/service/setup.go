package sphinx

import (
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/utils/httputil"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Service struct {
	httpApi   *httputil.HttpUtil
	appId     string
	appSecret string
	appToken  string
}

func New(baseUrl, appId, appSecret string) *Service {
	httpApi := httputil.New(baseUrl)
	appToken := base64.StdEncoding.EncodeToString([]byte(appId + ":" + appSecret))

	return &Service{
		httpApi:   httpApi,
		appId:     appId,
		appSecret: appSecret,
		appToken:  appToken,
	}
}

type Account struct {
	Id             uuid.UUID          `json:"id" validate:"required"`
	Email          htypes.Email       `json:"email" validate:"required"`
	Phone          htypes.PhoneNumber `json:"phone,omitempty"`
	Username       string             `json:"username,omitempty" validate:"wordId"`
	Document       htypes.Document    `json:"document,omitempty"`
	auth.ExtraData `json:"extraData,omitempty"`

	IsActive             bool       `json:"isActive"`
	HasEmailBeenVerified bool       `json:"hasEmailBeenVerified"`
	HasPhoneBeenVerified bool       `json:"hasPhoneBeenVerified"`
	Link                 *auth.Link `json:"link"`
}

func (a Account) IsAdmin() bool {
	for _, r := range a.Link.Roles {
		if r == auth.RoleValues.ADMIN {
			return true
		}
	}

	return false
}

func (a Account) IsStaff() bool {
	for _, r := range a.Link.Roles {
		if r == auth.RoleValues.STAFF {
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

func (a Account) HasGranting(granting string) bool {
	for _, g := range a.Link.Grantings {
		if g == granting {
			return true
		}
	}

	return false
}
