package oauthcase

import (
	"log"
	"time"

	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/cornucopia/utils/httputil"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type OAuthCallback struct {
	CacheRepo cache.DAO
}

type OAuthCallbackInput struct {
	SphinxApiBaseUrl string
	State            string
	Code             string
	AppId            string
	AppSecret        string
}

func (i OAuthCallback) Execute(input OAuthCallbackInput) (string, error) {
	key := OAuthStateKey(input.State)
	data := map[string]string{}
	err := i.CacheRepo.GetJson(key, &data)
	if err != nil {
		return "", err
	} else if len(data) == 0 {
		return "", normalizederr.NewRequestError("invalid state")
	}

	log.Printf("Origin Callback: %v", data["origin"])
	var output presenter.Success[authcase.LoginOutput]
	_, err = httputil.New(input.SphinxApiBaseUrl).Post("/auth/open/login", map[string]any{
		"code":      input.Code,
		"appSecret": input.AppSecret,
	}, &httputil.Options{
		Headers: map[string]string{
			"X-App": input.AppId,
		},
	})(&output)
	if err != nil {
		return "", err
	}

	data["accountId"] = output.Data.AccountId.String()
	data["accessToken"] = output.Data.AccessToken
	data["refreshToken"] = output.Data.RefreshToken

	err = i.CacheRepo.CacheJson(key, data, 5*time.Minute)
	if err != nil {
		return "", err
	}

	return data["origin"], nil
}
