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

func TestUserCreation(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.server.Close()

	factory := NewTestDataFactory()
	t.Run("should create a new user", func(t *testing.T) {
		userData := factory.FullUser()

		resp, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var respData presenter.Success[auth.User]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.Equal(t, userData["username"], respData.Data.Username)
		assert.Equal(t, userData["name"], respData.Data.ExtraData.Name)
		assert.Equal(t, userData["surname"], respData.Data.ExtraData.Surname)
		assert.True(t, respData.Data.IsActive)

		// Sensitive data should not be returned
		assert.Empty(t, respData.Data.Email)
		assert.Empty(t, respData.Data.Phone)
		assert.Empty(t, respData.Data.Document)
		assert.Empty(t, respData.Data.Password)
		assert.Zero(t, respData.Data.ExtraData.Address)

		user, err := ts.server.GetMockQueries().GetUserByID(respData.Data.ID)
		require.NoError(t, err)
		require.NotNil(t, user)

		// Assert not returned data
		assert.Equal(t, userData["email"], user.Email.String())
		assert.Equal(t, userData["phone"], user.Phone.String())
		assert.Equal(t, userData["document"], user.Document.String())
		dataAddress := userData["address"].(map[string]interface{})
		assert.Equal(t, dataAddress["line1"], user.ExtraData.Address.Line1)
		assert.Equal(t, dataAddress["number"], user.ExtraData.Address.Number)
		assert.Equal(t, dataAddress["city"], user.ExtraData.Address.City)
		assert.Equal(t, dataAddress["state"], user.ExtraData.Address.State)
		assert.Equal(t, dataAddress["country"], user.ExtraData.Address.Country.Name())
		assert.Equal(t, dataAddress["zipCode"], string(user.ExtraData.Address.ZipCode))
		assert.False(t, user.HasEmailBeenVerified)

		t.Run("should verify email", func(t *testing.T) {
			resp, err := ts.Request("PATCH", "/user/"+respData.Data.ID.String()+"/verification", map[string]string{
				"code": user.VerificationCodes[auth.VerificationEmail],
				"kind": string(auth.VerificationEmail),
			}, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)

			user, err := ts.server.GetMockQueries().GetUserByID(respData.Data.ID)
			require.NoError(t, err)
			require.NotNil(t, user)
			assert.True(t, user.HasEmailBeenVerified)
		})
	})

	t.Run("should normalize email, document and phone", func(t *testing.T) {
		userData := factory.RandomUser()
		userData["email"] = strings.ToUpper(userData["email"].(string))
		userData["document"] = GenerateCPF("XXX.XXX.XXX-XX")
		userData["phone"] = GeneratePhone("X (X) XXXX-XXXX")

		resp, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var respData presenter.Success[auth.User]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		user, err := ts.server.GetMockQueries().GetUserByID(respData.Data.ID)
		require.NoError(t, err)
		require.NotNil(t, user)

		// Assert not returned data
		expectedData := map[string]interface{}{
			"email":    strings.ToLower(userData["email"].(string)),
			"document": "cpf:" + sanitizer.Digit(userData["document"].(string)),
			"phone":    "+" + sanitizer.Digit(userData["phone"].(string)),
		}
		assert.Equal(t, expectedData["email"], user.Email.String())
		assert.Equal(t, expectedData["phone"], user.Phone.String())
		assert.Equal(t, expectedData["document"], user.Document.String())
	})

	t.Run("should reject invalid email", func(t *testing.T) {
		userData := map[string]interface{}{
			"email":    "invalid-email",
			"password": "TestPassword123!",
			"username": "testuser",
		}

		resp, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("should reject weak password", func(t *testing.T) {
		userData := map[string]interface{}{
			"email":    "test2@example.com",
			"password": "123", // Too weak
			"username": "testuser2",
		}

		resp, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("should reject duplicate email", func(t *testing.T) {
		userData := map[string]interface{}{
			"email":    "duplicate@example.com",
			"password": "TestPassword123!",
		}

		// Create first user
		resp1, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		resp1.Body.Close()
		assert.Equal(t, http.StatusCreated, resp1.StatusCode)

		// Try to create duplicate
		resp2, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusConflict, resp2.StatusCode)

		// Try to create duplicate with different case
		userData["email"] = "Duplicate@EXAMPLE.COM"
		resp3, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp3.Body.Close()
		assert.Equal(t, http.StatusConflict, resp3.StatusCode)
	})
}

func TestUserManagement(t *testing.T) {
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

	t.Run("should update user information", func(t *testing.T) {
		resp, err := ts.AuthenticatedRequest("PATCH", "/user", updateData, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accData presenter.Success[auth.UserPrivateView]
		err = json.NewDecoder(resp.Body).Decode(&accData)
		require.NoError(t, err)
		t.Log(accData)

		assert.Equal(t, updateData["name"], accData.Data.Name)
		assert.Equal(t, updateData["surname"], accData.Data.Surname)
	})

	t.Run("should update unique user information", func(t *testing.T) {
		resp, err := ts.AuthenticatedRequest("PATCH", "/user/unique", updateUniqueData, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accData presenter.Success[auth.User]
		err = json.NewDecoder(resp.Body).Decode(&accData)
		require.NoError(t, err)

		assert.Equal(t, updateUniqueData["username"], accData.Data.Username)
	})

	t.Run("should retrieve user information", func(t *testing.T) {
		resp, err := ts.AuthenticatedRequest("GET", "/user", nil, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accData presenter.Success[auth.UserPrivateView]
		err = json.NewDecoder(resp.Body).Decode(&accData)
		require.NoError(t, err)

		assert.Equal(t, updateUniqueData["username"], accData.Data.Username)
		assert.Equal(t, updateData["name"], accData.Data.Name)
		assert.Equal(t, updateData["surname"], accData.Data.Surname)
		assert.Equal(t, mocks.SimpleUserUser.Email.String(), accData.Data.Email.String())
		assert.Equal(t, mocks.SimpleUserUser.Phone.String(), accData.Data.Phone.String())
		assert.Equal(t, mocks.SimpleUserUser.Document.String(), accData.Data.Document.String())
	})
}

func TestPasswordChange(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	userData := factory.RandomUser()
	newPassword := "NewTestPassword123!"

	t.Run("should handle password change flow", func(t *testing.T) {
		// Create user
		resp1, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		resp1.Body.Close()

		// Login to get token
		loginData := factory.CreateLoginData(userData["email"].(string), userData["password"].(string))
		resp2, err := ts.Request("POST", "/auth/login", loginData, nil)
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusOK, resp2.StatusCode)
		var loginRespData presenter.Success[authcase.LoginOutput]
		err = json.NewDecoder(resp2.Body).Decode(&loginRespData)
		require.NoError(t, err)

		// Request password reset (assuming endpoint exists)
		changePassword := map[string]interface{}{
			"oldPassword": userData["password"],
			"newPassword": newPassword,
		}

		resp, err := ts.AuthenticatedRequest("PATCH", "/user/password", changePassword, loginRespData.Data.AccessToken)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		t.Run("should unauthorize login with old password", func(t *testing.T) {
			// Attempt login with old password
			loginData := factory.CreateLoginData(userData["email"].(string), userData["password"].(string))
			resp, err := ts.Request("POST", "/auth/login", loginData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("should login with new password", func(t *testing.T) {
			// Attempt login with new password
			loginData := factory.CreateLoginData(userData["email"].(string), newPassword)
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
	userData := factory.RandomUser()
	newPassword := "NewPassword123!"

	t.Run("should handle password reset flow", func(t *testing.T) {
		// Create user
		resp1, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp1.Body.Close()

		var accData presenter.Success[auth.User]
		err = json.NewDecoder(resp1.Body).Decode(&accData)
		require.NoError(t, err)

		// Request password reset
		resetRequest := map[string]interface{}{
			"entry": userData["email"],
		}

		resp, err := ts.Request("POST", "/user/password/request", resetRequest, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		t.Run("should reset password", func(t *testing.T) {
			user, err := ts.server.GetMockQueries().GetUserByID(accData.Data.ID)
			require.NoError(t, err)
			require.NotNil(t, user)

			resetData := map[string]interface{}{
				"code":        user.VerificationCodes[auth.VerificationPasswordReset],
				"newPassword": newPassword,
			}

			resp, err := ts.Request("PATCH", "/user/"+user.ID.String()+"/password", resetData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)

			t.Run("should login with new password", func(t *testing.T) {
				loginData := factory.CreateLoginData(userData["email"].(string), newPassword)
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
				loginData := factory.CreateLoginData(userData["email"].(string), userData["password"].(string))
				resp, err := ts.Request("POST", "/auth/login", loginData, nil)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			})
		})

	})
}

func TestUserExistence(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	t.Run("should handle different entry types, even not normalized", func(t *testing.T) {
		factory := NewTestDataFactory()
		userData := factory.FullUser()

		resp, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		t.Run("should check existence by username", func(t *testing.T) {
			// Username as-is
			existsResp1, err := ts.Request("GET", "/user/existence", userData, map[string]string{
				"x-entry": userData["username"].(string),
			})
			require.NoError(t, err)
			defer existsResp1.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp1.StatusCode)

			var existsData1 presenter.Success[bool]
			err = json.NewDecoder(existsResp1.Body).Decode(&existsData1)
			require.NoError(t, err)

			assert.True(t, existsData1.Data, "User should exist for username entry")

			// Username in upper case (case insensitive check)
			newFormatUsername := strings.ToUpper(userData["username"].(string))

			existsResp2, err := ts.Request("GET", "/user/existence", userData, map[string]string{
				"x-entry": newFormatUsername,
			})
			require.NoError(t, err)
			defer existsResp2.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp2.StatusCode)

			var existsData2 presenter.Success[bool]
			err = json.NewDecoder(existsResp2.Body).Decode(&existsData2)
			require.NoError(t, err)

			assert.True(t, existsData2.Data, "User should exist for username in upper case entry")
		})

		t.Run("should check existence by email", func(t *testing.T) {
			// Email as-is
			existsResp1, err := ts.Request("GET", "/user/existence", userData, map[string]string{
				"x-entry": userData["email"].(string),
			})
			require.NoError(t, err)
			defer existsResp1.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp1.StatusCode)

			var existsData1 presenter.Success[bool]
			err = json.NewDecoder(existsResp1.Body).Decode(&existsData1)
			require.NoError(t, err)

			assert.True(t, existsData1.Data, "User should exist for email entry")

			// Email in upper case (case insensitive check)
			newFormatEmail := strings.ToUpper(userData["email"].(string))

			existsResp2, err := ts.Request("GET", "/user/existence", userData, map[string]string{
				"x-entry": newFormatEmail,
			})
			require.NoError(t, err)
			defer existsResp2.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp2.StatusCode)

			var existsData2 presenter.Success[bool]
			err = json.NewDecoder(existsResp2.Body).Decode(&existsData2)
			require.NoError(t, err)

			assert.True(t, existsData2.Data, "User should exist for email in upper case entry")
		})

		t.Run("should check existence by phone", func(t *testing.T) {
			// Phone as-is
			existsResp1, err := ts.Request("GET", "/user/existence", userData, map[string]string{
				"x-entry": userData["phone"].(string),
			})
			require.NoError(t, err)
			defer existsResp1.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp1.StatusCode)

			var existsData1 presenter.Success[bool]
			err = json.NewDecoder(existsResp1.Body).Decode(&existsData1)
			require.NoError(t, err)

			assert.True(t, existsData1.Data, "User should exist for phone entry")

			// Phone in different format
			newFormatPhone := sanitizer.Digit(userData["phone"].(string))
			newFormatPhone = "+" + newFormatPhone[:3] + " (" + newFormatPhone[3:5] + ")" + newFormatPhone[5:]

			existsResp2, err := ts.Request("GET", "/user/existence", userData, map[string]string{
				"x-entry": newFormatPhone,
			})
			require.NoError(t, err)
			defer existsResp2.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp2.StatusCode)

			var existsData2 presenter.Success[bool]
			err = json.NewDecoder(existsResp2.Body).Decode(&existsData2)
			require.NoError(t, err)

			assert.True(t, existsData2.Data, "User should exist for phone in different format")
		})

		t.Run("should check existence by document", func(t *testing.T) {
			// Document as-is
			existsResp1, err := ts.Request("GET", "/user/existence", userData, map[string]string{
				"x-entry": userData["document"].(string),
			})
			require.NoError(t, err)
			defer existsResp1.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp1.StatusCode)

			var existsData1 presenter.Success[bool]
			err = json.NewDecoder(existsResp1.Body).Decode(&existsData1)
			require.NoError(t, err)

			assert.True(t, existsData1.Data, "User should exist for document entry")

			// Document in only digit format
			newFormatDocument := sanitizer.Digit(userData["document"].(string))

			existsResp2, err := ts.Request("GET", "/user/existence", userData, map[string]string{
				"x-entry": newFormatDocument,
			})
			require.NoError(t, err)
			defer existsResp2.Body.Close()

			assert.Equal(t, http.StatusOK, existsResp2.StatusCode)

			var existsData2 presenter.Success[bool]
			err = json.NewDecoder(existsResp2.Body).Decode(&existsData2)
			require.NoError(t, err)

			assert.True(t, existsData2.Data, "User should exist for document in only digit format")
		})
	})
}
