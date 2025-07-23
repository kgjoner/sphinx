// Test the most simple authentication flow with real server, i.e, real external integrations like postgres and redis.
package server_test

import (
	"net/http/httptest"
	"testing"

	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/utils/httputil"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
	"github.com/kgjoner/sphinx/internal/server"
	"github.com/stretchr/testify/assert"
)

func startTestServer() *httptest.Server {
	config.Must()
	return httptest.NewServer(server.New().Setup().Handler)
}

var (
	unhashedPassword = "test123!"
	mockedAccount    = &auth.Account{
		Email:    "test@test.com",
		Username: "test",
		Phone:    "+5511999999999",
		Document: "cpf:02496946031",
		Password: unhashedPassword,
		IsActive: true,
	}
)

func TestAccount(t *testing.T) {
	s := startTestServer()
	defer s.Close()

	api := httputil.New(s.URL + "/v1")

	t.Run("it should create an account", func(t *testing.T) {
		var respData presenter.Success[auth.Account]
		resp, err := api.Post("/account", map[string]any{
			"email":    mockedAccount.Email.String(),
			"password": unhashedPassword,
			"username": mockedAccount.Username,
			"phone":    mockedAccount.Phone,
			"document": mockedAccount.Document,
		}, &httputil.Options{
			Headers: map[string]string{
				"x-app": config.Env.ROOT_APP_ID,
			},
		})(&respData)

		assert.Nil(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		assert.Equal(t, mockedAccount.Username, respData.Data.Username)
		assert.Equal(t, mockedAccount.IsActive, respData.Data.IsActive)
	})

	t.Run("it should log in", func(t *testing.T) {
		var respData presenter.Success[authcase.LoginOutput]
		resp, err := api.Post("/auth/login", map[string]any{
			"entry":    mockedAccount.Email.String(),
			"password": unhashedPassword,
		}, &httputil.Options{
			Headers: map[string]string{
				"x-app": config.Env.ROOT_APP_ID,
			},
		})(&respData)

		currentToken := respData.Data.AccessToken

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.NotZero(t, respData.Data.RefreshToken)
		assert.NotZero(t, respData.Data.AccessToken)
		assert.NotZero(t, respData.Data.AccountId)

		t.Run("it should retrieve user info", func(t *testing.T) {
			var respData presenter.Success[auth.AccountPrivateView]
			resp, err := api.Get("/account", &httputil.Options{
				Headers: map[string]string{
					"authorization": "Bearer " + currentToken,
				},
			})(&respData)

			assert.Nil(t, err)
			assert.Equal(t, 200, resp.StatusCode)
			assert.Equal(t, mockedAccount.Email, respData.Data.Email)
			assert.Equal(t, mockedAccount.Phone, respData.Data.Phone)
			assert.Equal(t, mockedAccount.Username, respData.Data.Username)
			assert.Equal(t, mockedAccount.Document, respData.Data.Document)
		})
	})
}
