package oauthcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/repositories/cache"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type RetrieveToken struct {
	CacheRepo cache.DAO
}

type RetrieveTokenInput struct {
	SessionId string
	State     string
	CsrfToken string
}

func (i RetrieveToken) Execute(input RetrieveTokenInput) (*authcase.LoginViaOAuthOutput, error) {
	key := OAuthStateKey(input.State)
	data := map[string]string{}
	err := i.CacheRepo.GetJson(key, data)
	if err != nil {
		return nil, err
	} else if len(data) == 0 {
		return nil, normalizederr.NewRequestError("invalid state")
	}

	if data["state"] != input.State || data["sessionId"] != input.SessionId || data["csrfToken"] != input.CsrfToken {
		return nil, normalizederr.NewFatalUnauthorizedError("invalid state, session id or csrf token")
	}

	accountId, err := uuid.Parse(data["accountId"])
	if err != nil {
		return nil, err
	}

	return &authcase.LoginViaOAuthOutput{
		AccountId:    accountId,
		AccessToken:  data["accessToken"],
		RefreshToken: data["refreshToken"],
	}, nil
}
