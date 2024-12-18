package oauthcase

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	cacherepo "github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/cornucopia/utils/pwdgen"
)

func OAuthStateKey(state string) string {
	return "oauth:" + state
}

type StartOAuth struct {
	CacheRepo cacherepo.Queries
}

type StartOAuthInput struct {
	SphinxClientBaseUrl string
	AppBaseUrl          string
	AppId               string
}

func (i StartOAuth) Execute(input StartOAuthInput) (*StartOAuthOutput, *http.Cookie, error) {
	data := map[string]string{
		"state":     pwdgen.Generate(24, "lower", "upper", "number"),
		"sessionId": pwdgen.Generate(24, "lower", "upper", "number"),
		"csrfToken": pwdgen.Generate(24, "lower", "upper", "number"),
	}

	redirectUri := fmt.Sprintf("%v/oauth/callback",
		input.AppBaseUrl,
	)
	authorizationUrl := fmt.Sprintf("%v?client_id=%v&redirect_uri=%v&state=%v",
		input.SphinxClientBaseUrl,
		input.AppId,
		url.QueryEscape(redirectUri),
		data["state"],
	)

	expirationTime := 1 * time.Hour
	expiresAt := time.Now().Add(expirationTime)
	err := i.CacheRepo.CacheJson(OAuthStateKey(data["state"]), data, expirationTime)
	if err != nil {
		return nil, nil, err
	}

	return &StartOAuthOutput{
			AuthorizationUrl: authorizationUrl,
			State:            data["state"],
			CsrfToken:        data["csrfToken"],
			ExpiresAt:        expiresAt,
		}, &http.Cookie{
			Name:     "session_id",
			Value:    data["sessionId"],
			Path:     "/oauth",
			Expires:  expiresAt,
			Secure:   true,
			HttpOnly: true,
		}, nil
}

type StartOAuthOutput struct {
	AuthorizationUrl string
	State            string
	CsrfToken        string
	ExpiresAt        time.Time
}
