package sphinx

import (
	"encoding/base64"

	"github.com/kgjoner/cornucopia/utils/httputil"
)

type SphinxService struct {
	httpApi   *httputil.HttpUtil
	appId     string
	appSecret string
	appToken  string
}

func New(baseUrl, appId, appSecret string) *SphinxService {
	httpApi := httputil.New(baseUrl)
	appToken := base64.StdEncoding.EncodeToString([]byte(appId + ":" + appSecret))

	return &SphinxService{
		httpApi:   httpApi,
		appId:     appId,
		appSecret: appSecret,
		appToken:  appToken,
	}
}
