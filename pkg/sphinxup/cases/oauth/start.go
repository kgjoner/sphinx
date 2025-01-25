package oauthcase

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/cornucopia/utils/pwdgen"
)

func OAuthStateKey(state string) string {
	return "oauth:" + state
}

type StartOAuth struct {
	CacheRepo cache.DAO
}

type StartOAuthInput struct {
	Origin              string
	SphinxClientBaseUrl string
	AppDomain           string
	AppBasePath         string
	AppId               string
}

func (i StartOAuth) Execute(input StartOAuthInput) (*StartOAuthOutput, *http.Cookie, error) {
	data := map[string]string{
		"state":     pwdgen.Generate(24, "lower", "upper", "number"),
		"sessionId": pwdgen.Generate(24, "lower", "upper", "number"),
		"csrfToken": pwdgen.Generate(24, "lower", "upper", "number"),
		"origin":    input.Origin,
	}

	redirectUri := fmt.Sprintf("%v%v/oauth/callback",
		input.AppDomain, input.AppBasePath,
	)
	authorizationUrl := fmt.Sprintf("%v/%v?path=oauth&redirect_uri=%v&state=%v",
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

	cookie := http.Cookie{
		Name:     "session_id",
		Value:    data["sessionId"],
		Path:     input.AppBasePath + "/oauth",
		Expires:  expiresAt,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}

	return &StartOAuthOutput{
		AuthorizationUrl: authorizationUrl,
		State:            data["state"],
		CsrfToken:        data["csrfToken"],
		ExpiresAt:        expiresAt,
	}, &cookie, nil
}

type StartOAuthOutput struct {
	AuthorizationUrl string    `json:"authorizationUrl"`
	State            string    `json:"state"`
	CsrfToken        string    `json:"csrfToken"`
	ExpiresAt        time.Time `json:"expiresAt"`
}
