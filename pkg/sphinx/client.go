package sphinx

import (
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/utils/httputil"
)

type Client struct {
	httpApi   *httputil.HTTPUtil
	baseURL   string
	appID     string
	appSecret string
	appToken  string
}

func NewClient(baseURL, appID, appSecret string) *Client {
	httpApi := httputil.New(baseURL)
	appToken := base64.StdEncoding.EncodeToString([]byte(appID + ":" + appSecret))

	svc := &Client{
		httpApi:   httpApi,
		baseURL:   baseURL,
		appID:     appID,
		appSecret: appSecret,
		appToken:  appToken,
	}

	return svc
}

// Middlewares creates HTTP middlewares for token authentication using JWKS.
// This method provides a convenient way to create middlewares that automatically
// fetch and validate JWT tokens using the Sphinx server's public keys.
//
// Example:
//
//	client := sphinx.NewClient("https://sphinx.example.com", appID, appSecret)
//	middlewares := client.Middlewares(authorizer)
//	router.Use(middlewares.Authenticate)
func (c *Client) Middlewares(authorizer Authorizer) AuthMiddlewares {
	appIDUUID, err := uuid.Parse(c.appID)
	if err != nil {
		// Fallback to zero UUID if parsing fails
		appIDUUID = uuid.Nil
	}

	return NewMiddlewares(appIDUUID, c.baseURL, authorizer)
}
