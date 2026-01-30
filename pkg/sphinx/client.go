package sphinx

import (
	"encoding/base64"

	"github.com/kgjoner/cornucopia/v2/utils/httputil"
)

type Client struct {
	httpApi   *httputil.HTTPUtil
	appID     string
	appSecret string
	appToken  string
}

func NewClient(baseURL, appID, appSecret string) *Client {
	httpApi := httputil.New(baseURL)
	appToken := base64.StdEncoding.EncodeToString([]byte(appID + ":" + appSecret))

	svc := &Client{
		httpApi:   httpApi,
		appID:     appID,
		appSecret: appSecret,
		appToken:  appToken,
	}

	return svc
}
