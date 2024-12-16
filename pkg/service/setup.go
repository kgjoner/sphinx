package sphinx

import (
	"encoding/base64"

	"github.com/kgjoner/cornucopia/utils/httputil"
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
