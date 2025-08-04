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
	mockedUser       = &auth.User{
		Email:    "test@test.com",
		Username: "test",
		Phone:    "+5511999999999",
		Document: "cpf:02496946031",
		Password: unhashedPassword,
		IsActive: true,
	}
)

func TestUser(t *testing.T) {
	s := startTestServer()
	defer s.Close()

	api := httputil.New(s.URL + "/v1")

	t.Run("it should create an user", func(t *testing.T) {
		var respData presenter.Success[auth.User]
		resp, err := api.Post("/user", map[string]any{
			"email":    mockedUser.Email.String(),
			"password": unhashedPassword,
			"username": mockedUser.Username,
			"phone":    mockedUser.Phone,
			"document": mockedUser.Document,
		}, nil)(&respData)

		assert.Nil(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		assert.Equal(t, mockedUser.Username, respData.Data.Username)
		assert.Equal(t, mockedUser.IsActive, respData.Data.IsActive)
	})

	t.Run("it should log in", func(t *testing.T) {
		var respData presenter.Success[authcase.LoginOutput]
		resp, err := api.Post("/auth/login", map[string]any{
			"entry":    mockedUser.Email.String(),
			"password": unhashedPassword,
		}, nil)(&respData)

		currentToken := respData.Data.AccessToken

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.NotZero(t, respData.Data.RefreshToken)
		assert.NotZero(t, respData.Data.AccessToken)
		assert.NotZero(t, respData.Data.UserID)

		t.Run("it should retrieve user info", func(t *testing.T) {
			var respData presenter.Success[auth.UserPrivateView]
			resp, err := api.Get("/user", &httputil.Options{
				Headers: map[string]string{
					"authorization": "Bearer " + currentToken,
				},
			})(&respData)

			assert.Nil(t, err)
			assert.Equal(t, 200, resp.StatusCode)
			assert.Equal(t, mockedUser.Email, respData.Data.Email)
			assert.Equal(t, mockedUser.Phone, respData.Data.Phone)
			assert.Equal(t, mockedUser.Username, respData.Data.Username)
			assert.Equal(t, mockedUser.Document, respData.Data.Document)
		})
	})
}
