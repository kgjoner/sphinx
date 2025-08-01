package e2e

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/utils/sanitizer"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
	"github.com/kgjoner/sphinx/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountCreation(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.server.Close()

	factory := NewTestDataFactory()
	t.Run("should create a new account", func(t *testing.T) {
		accountData := factory.FullAccount()

		resp, err := ts.Request("POST", "/account", accountData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var respData presenter.Success[auth.Account]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.Equal(t, accountData["username"], respData.Data.Username)
		assert.Equal(t, accountData["name"], respData.Data.ExtraData.Name)
		assert.Equal(t, accountData["surname"], respData.Data.ExtraData.Surname)
		assert.True(t, respData.Data.IsActive)

		// Sensitive data should not be returned
		assert.Empty(t, respData.Data.Email)
		assert.Empty(t, respData.Data.Phone)
		assert.Empty(t, respData.Data.Document)
		assert.Empty(t, respData.Data.Password)
		assert.Zero(t, respData.Data.ExtraData.Address)

		acc, err := ts.server.GetMockQueries().GetAccountByID(respData.Data.ID)
		require.NoError(t, err)
		require.NotNil(t, acc)

		// Assert not returned data
		assert.Equal(t, accountData["email"], acc.Email.String())
		assert.Equal(t, accountData["phone"], acc.Phone.String())
		assert.Equal(t, accountData["document"], acc.Document.String())
		dataAddress := accountData["address"].(map[string]interface{})
		assert.Equal(t, dataAddress["line1"], acc.ExtraData.Address.Line1)
		assert.Equal(t, dataAddress["number"], acc.ExtraData.Address.Number)
		assert.Equal(t, dataAddress["city"], acc.ExtraData.Address.City)
		assert.Equal(t, dataAddress["state"], acc.ExtraData.Address.State)
		assert.Equal(t, dataAddress["country"], acc.ExtraData.Address.Country.Name())
		assert.Equal(t, dataAddress["zipCode"], string(acc.ExtraData.Address.ZipCode))
		assert.False(t, acc.HasEmailBeenVerified)

		t.Run("should verify email", func(t *testing.T) {
			resp, err := ts.Request("PATCH", "/account/"+respData.Data.ID.String()+"/verification", map[string]string{
				"code": acc.VerificationCodes[auth.VerificationEmail],
				"kind": string(auth.VerificationEmail),
			}, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)

			acc, err := ts.server.GetMockQueries().GetAccountByID(respData.Data.ID)
			require.NoError(t, err)
			require.NotNil(t, acc)
			assert.True(t, acc.HasEmailBeenVerified)
		})
	})

	t.Run("should normalize email, document and phone", func(t *testing.T) {
		accountData := factory.RandomAccount()
		accountData["email"] = strings.ToUpper(accountData["email"].(string))
		accountData["document"] = GenerateCPF("XXX.XXX.XXX-XX")
		accountData["phone"] = GeneratePhone("X (X) XXXX-XXXX")

		resp, err := ts.Request("POST", "/account", accountData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var respData presenter.Success[auth.Account]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		acc, err := ts.server.GetMockQueries().GetAccountByID(respData.Data.ID)
		require.NoError(t, err)
		require.NotNil(t, acc)

		// Assert not returned data
		expectedData := map[string]interface{}{
			"email":    strings.ToLower(accountData["email"].(string)),
			"document": "cpf:" + sanitizer.Digit(accountData["document"].(string)),
			"phone":    "+" + sanitizer.Digit(accountData["phone"].(string)),
		}
		assert.Equal(t, expectedData["email"], acc.Email.String())
		assert.Equal(t, expectedData["phone"], acc.Phone.String())
		assert.Equal(t, expectedData["document"], acc.Document.String())
	})

	t.Run("should reject invalid email", func(t *testing.T) {
		accountData := map[string]interface{}{
			"email":    "invalid-email",
			"password": "TestPassword123!",
			"username": "testuser",
		}

		resp, err := ts.Request("POST", "/account", accountData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("should reject weak password", func(t *testing.T) {
		accountData := map[string]interface{}{
			"email":    "test2@example.com",
			"password": "123", // Too weak
			"username": "testuser2",
		}

		resp, err := ts.Request("POST", "/account", accountData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("should reject duplicate email", func(t *testing.T) {
		accountData := map[string]interface{}{
			"email":    "duplicate@example.com",
			"password": "TestPassword123!",
		}

		// Create first account
		resp1, err := ts.Request("POST", "/account", accountData, nil)
		require.NoError(t, err)
		resp1.Body.Close()
		assert.Equal(t, http.StatusCreated, resp1.StatusCode)

		// Try to create duplicate
		resp2, err := ts.Request("POST", "/account", accountData, nil)
		require.NoError(t, err)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusConflict, resp2.StatusCode)

		// Try to create duplicate with different case
		accountData["email"] = "Duplicate@EXAMPLE.COM"
		resp3, err := ts.Request("POST", "/account", accountData, nil)
		require.NoError(t, err)
		defer resp3.Body.Close()
		assert.Equal(t, http.StatusConflict, resp3.StatusCode)
	})
}

func TestAccountManagement(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	// Login
	loginResp, err := ts.Request("POST", "/auth/login", factory.SimpleUserLoginData(), nil)
	require.NoError(t, err)
	defer loginResp.Body.Close()

	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	token := loginResult["data"].(map[string]interface{})["accessToken"].(string)

	// Updated data
	updateData := map[string]interface{}{
		"name":    "Updated User",
		"surname": "Updated Surname",
	}

	updateUniqueData := map[string]interface{}{
		"username": "updated.username",
	}

	t.Run("should update account information", func(t *testing.T) {
		resp, err := ts.AuthenticatedRequest("PATCH", "/account", updateData, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accData presenter.Success[auth.AccountPrivateView]
		err = json.NewDecoder(resp.Body).Decode(&accData)
		require.NoError(t, err)
		t.Log(accData)

		assert.Equal(t, updateData["name"], accData.Data.Name)
		assert.Equal(t, updateData["surname"], accData.Data.Surname)
	})

	t.Run("should update unique account information", func(t *testing.T) {
		resp, err := ts.AuthenticatedRequest("PATCH", "/account/unique", updateUniqueData, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accData presenter.Success[auth.Account]
		err = json.NewDecoder(resp.Body).Decode(&accData)
		require.NoError(t, err)

		assert.Equal(t, updateUniqueData["username"], accData.Data.Username)
	})

	t.Run("should retrieve account information", func(t *testing.T) {
		resp, err := ts.AuthenticatedRequest("GET", "/account", nil, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accData presenter.Success[auth.AccountPrivateView]
		err = json.NewDecoder(resp.Body).Decode(&accData)
		require.NoError(t, err)

		assert.Equal(t, updateUniqueData["username"], accData.Data.Username)
		assert.Equal(t, updateData["name"], accData.Data.Name)
		assert.Equal(t, updateData["surname"], accData.Data.Surname)
		assert.Equal(t, mocks.SimpleUserAccount.Email.String(), accData.Data.Email.String())
		assert.Equal(t, mocks.SimpleUserAccount.Phone.String(), accData.Data.Phone.String())
		assert.Equal(t, mocks.SimpleUserAccount.Document.String(), accData.Data.Document.String())
	})
}

func TestPasswordChange(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	accountData := factory.RandomAccount()
	newPassword := "NewTestPassword123!"

	t.Run("should handle password change flow", func(t *testing.T) {
		// Create account
		resp1, err := ts.Request("POST", "/account", accountData, nil)
		require.NoError(t, err)
		resp1.Body.Close()

		// Login to get token
		loginData := factory.CreateLoginData(accountData["email"].(string), accountData["password"].(string))
		resp2, err := ts.Request("POST", "/auth/login", loginData, nil)
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusOK, resp2.StatusCode)
		var loginRespData presenter.Success[authcase.LoginOutput]
		err = json.NewDecoder(resp2.Body).Decode(&loginRespData)
		require.NoError(t, err)

		// Request password reset (assuming endpoint exists)
		changePassword := map[string]interface{}{
			"oldPassword": accountData["password"],
			"newPassword": newPassword,
		}

		resp, err := ts.AuthenticatedRequest("PATCH", "/account/password", changePassword, loginRespData.Data.AccessToken)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		t.Run("should unauthorize login with old password", func(t *testing.T) {
			// Attempt login with old password
			loginData := factory.CreateLoginData(accountData["email"].(string), accountData["password"].(string))
			resp, err := ts.Request("POST", "/auth/login", loginData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("should login with new password", func(t *testing.T) {
			// Attempt login with new password
			loginData := factory.CreateLoginData(accountData["email"].(string), newPassword)
			resp, err := ts.Request("POST", "/auth/login", loginData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var loginRespData presenter.Success[authcase.LoginOutput]
			err = json.NewDecoder(resp.Body).Decode(&loginRespData)
			require.NoError(t, err)

			assert.NotEmpty(t, loginRespData.Data.AccessToken)
		})
	})
}

func TestPasswordReset(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()
	accountData := factory.RandomAccount()
	newPassword := "NewPassword123!"

	t.Run("should handle password reset flow", func(t *testing.T) {
		// Create account
		resp1, err := ts.Request("POST", "/account", accountData, nil)
		require.NoError(t, err)
		defer resp1.Body.Close()

		var accData presenter.Success[auth.Account]
		err = json.NewDecoder(resp1.Body).Decode(&accData)
		require.NoError(t, err)

		// Request password reset
		resetRequest := map[string]interface{}{
			"entry": accountData["email"],
		}

		resp, err := ts.Request("POST", "/account/password/request", resetRequest, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		t.Run("should reset password", func(t *testing.T) {
			acc, err := ts.server.GetMockQueries().GetAccountByID(accData.Data.ID)
			require.NoError(t, err)
			require.NotNil(t, acc)

			resetData := map[string]interface{}{
				"code":        acc.VerificationCodes[auth.VerificationPasswordReset],
				"newPassword": newPassword,
			}

			resp, err := ts.Request("PATCH", "/account/"+acc.ID.String()+"/password", resetData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)

			t.Run("should login with new password", func(t *testing.T) {
				loginData := factory.CreateLoginData(accountData["email"].(string), newPassword)
				resp, err := ts.Request("POST", "/auth/login", loginData, nil)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)

				var loginRespData presenter.Success[authcase.LoginOutput]
				err = json.NewDecoder(resp.Body).Decode(&loginRespData)
				require.NoError(t, err)

				assert.NotEmpty(t, loginRespData.Data.AccessToken)
			})

			t.Run("should reject login with old password", func(t *testing.T) {
				// Attempt login with old password
				loginData := factory.CreateLoginData(accountData["email"].(string), accountData["password"].(string))
				resp, err := ts.Request("POST", "/auth/login", loginData, nil)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			})
		})

	})
}

func TestAccountExistence(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	t.Run("should handle different entry types, even not normalized", func(t *testing.T) {
		factory := NewTestDataFactory()
		accountData := factory.FullAccount()

		resp, err := ts.Request("POST", "/account", accountData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		t.Run("should check existence by username", func(t *testing.T) {
			// Username as-is
			existsResp1, err := ts.Request("GET", "/account/existence", accountData, map[string]string{
				"x-entry": accountData["username"].(string),
			})
			require.NoError(t, err)
			defer existsResp1.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp1.StatusCode)

			var existsData1 presenter.Success[bool]
			err = json.NewDecoder(existsResp1.Body).Decode(&existsData1)
			require.NoError(t, err)

			assert.True(t, existsData1.Data, "Account should exist for username entry")

			// Username in upper case (case insensitive check)
			newFormatUsername := strings.ToUpper(accountData["username"].(string))

			existsResp2, err := ts.Request("GET", "/account/existence", accountData, map[string]string{
				"x-entry": newFormatUsername,
			})
			require.NoError(t, err)
			defer existsResp2.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp2.StatusCode)

			var existsData2 presenter.Success[bool]
			err = json.NewDecoder(existsResp2.Body).Decode(&existsData2)
			require.NoError(t, err)

			assert.True(t, existsData2.Data, "Account should exist for username in upper case entry")
		})

		t.Run("should check existence by email", func(t *testing.T) {
			// Email as-is
			existsResp1, err := ts.Request("GET", "/account/existence", accountData, map[string]string{
				"x-entry": accountData["email"].(string),
			})
			require.NoError(t, err)
			defer existsResp1.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp1.StatusCode)

			var existsData1 presenter.Success[bool]
			err = json.NewDecoder(existsResp1.Body).Decode(&existsData1)
			require.NoError(t, err)

			assert.True(t, existsData1.Data, "Account should exist for email entry")

			// Email in upper case (case insensitive check)
			newFormatEmail := strings.ToUpper(accountData["email"].(string))

			existsResp2, err := ts.Request("GET", "/account/existence", accountData, map[string]string{
				"x-entry": newFormatEmail,
			})
			require.NoError(t, err)
			defer existsResp2.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp2.StatusCode)

			var existsData2 presenter.Success[bool]
			err = json.NewDecoder(existsResp2.Body).Decode(&existsData2)
			require.NoError(t, err)

			assert.True(t, existsData2.Data, "Account should exist for email in upper case entry")
		})

		t.Run("should check existence by phone", func(t *testing.T) {
			// Phone as-is
			existsResp1, err := ts.Request("GET", "/account/existence", accountData, map[string]string{
				"x-entry": accountData["phone"].(string),
			})
			require.NoError(t, err)
			defer existsResp1.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp1.StatusCode)

			var existsData1 presenter.Success[bool]
			err = json.NewDecoder(existsResp1.Body).Decode(&existsData1)
			require.NoError(t, err)

			assert.True(t, existsData1.Data, "Account should exist for phone entry")

			// Phone in different format
			newFormatPhone := sanitizer.Digit(accountData["phone"].(string))
			newFormatPhone = "+" + newFormatPhone[:3] + " (" + newFormatPhone[3:5] + ")" + newFormatPhone[5:]

			existsResp2, err := ts.Request("GET", "/account/existence", accountData, map[string]string{
				"x-entry": newFormatPhone,
			})
			require.NoError(t, err)
			defer existsResp2.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp2.StatusCode)

			var existsData2 presenter.Success[bool]
			err = json.NewDecoder(existsResp2.Body).Decode(&existsData2)
			require.NoError(t, err)

			assert.True(t, existsData2.Data, "Account should exist for phone in different format")
		})

		t.Run("should check existence by document", func(t *testing.T) {
			// Document as-is
			existsResp1, err := ts.Request("GET", "/account/existence", accountData, map[string]string{
				"x-entry": accountData["document"].(string),
			})
			require.NoError(t, err)
			defer existsResp1.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp1.StatusCode)

			var existsData1 presenter.Success[bool]
			err = json.NewDecoder(existsResp1.Body).Decode(&existsData1)
			require.NoError(t, err)

			assert.True(t, existsData1.Data, "Account should exist for document entry")

			// Document in only digit format
			newFormatDocument := sanitizer.Digit(accountData["document"].(string))

			existsResp2, err := ts.Request("GET", "/account/existence", accountData, map[string]string{
				"x-entry": newFormatDocument,
			})
			require.NoError(t, err)
			defer existsResp2.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp2.StatusCode)

			var existsData2 presenter.Success[bool]
			err = json.NewDecoder(existsResp2.Body).Decode(&existsData2)
			require.NoError(t, err)

			assert.True(t, existsData2.Data, "Account should exist for document in only digit format")
		})
	})
}
