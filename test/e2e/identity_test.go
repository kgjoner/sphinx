package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/cornucopia/v2/utils/sanitizer"
	"github.com/kgjoner/sphinx/internal/domains/auth/authcase"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/domains/identity/identrepo"
	"github.com/kgjoner/sphinx/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCreation(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.server.Close()

	factory := NewTestDataFactory()
	t.Run("should create a new user", func(t *testing.T) {
		userData := factory.RandomFullUser()

		resp, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var respData presenter.Success[identity.UserLeanView]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.Equal(t, userData["username"], respData.Data.Username)
		assert.Equal(t, userData["name"], respData.Data.Name)
		assert.Equal(t, userData["surname"], respData.Data.Surname)
		assert.True(t, respData.Data.IsActive)

		// Query the database to verify user was created
		dao := identrepo.NewFactory().NewDAO(context.Background(), ts.server.GetBasePool().Connection())
		user, err := dao.GetUserByID(respData.Data.ID)
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
			resp, err := ts.Request("POST", "/user/"+respData.Data.ID.String()+"/email/verification", map[string]string{
				"code": user.VerificationCodes[identity.VerificationEmail],
				"kind": string(identity.VerificationEmail),
			}, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)

			// Verify email was verified in database
			dao := identrepo.NewFactory().NewDAO(context.Background(), ts.server.GetBasePool().Connection())
			user, err := dao.GetUserByID(respData.Data.ID)
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

		var respData presenter.Success[identity.UserLeanView]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		// Query the database to verify user data
		dao := identrepo.NewFactory().NewDAO(context.Background(), ts.server.GetBasePool().Connection())
		user, err := dao.GetUserByID(respData.Data.ID)
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
		}

		resp, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("should reject weak password", func(t *testing.T) {
		userData := map[string]interface{}{
			"email":    GenerateEmail(),
			"password": "123", // Too weak
		}

		resp, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("should reject duplicate email", func(t *testing.T) {
		userData := factory.RandomUser()

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
		userData["email"] = strings.ToUpper(userData["email"].(string))
		resp3, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		defer resp3.Body.Close()
		assert.Equal(t, http.StatusConflict, resp3.StatusCode)
	})
}

func TestUserManagement(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	t.Run("should update user information", func(t *testing.T) {
		userData := ts.newUser(t)
		token := ts.login(t, userData.LoginPayload())

		// Updated data
		updateData := map[string]interface{}{
			"name":    "Updated User",
			"surname": "Updated Surname",
		}

		resp, err := ts.AuthenticatedRequest("PATCH", "/user/me", updateData, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accData presenter.Success[identity.UserView]
		err = json.NewDecoder(resp.Body).Decode(&accData)
		require.NoError(t, err)
		t.Log(accData)

		assert.Equal(t, updateData["name"], accData.Data.Name)
		assert.Equal(t, updateData["surname"], accData.Data.Surname)

		t.Run("should retrieve updated information", func(t *testing.T) {
			resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, token)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var accData presenter.Success[identity.UserView]
			err = json.NewDecoder(resp.Body).Decode(&accData)
			require.NoError(t, err)

			assert.Equal(t, updateData["name"], accData.Data.Name)
			assert.Equal(t, updateData["surname"], accData.Data.Surname)

			// Verify that email remains unchanged
			assert.Equal(t, userData.Email, accData.Data.Email.String())
		})
	})

	t.Run("should update unique user information", func(t *testing.T) {
		userData := ts.newUser(t)
		token := ts.login(t, userData.LoginPayload())

		updateUsernameData := map[string]interface{}{
			"value": "updated.username",
		}

		resp, err := ts.AuthenticatedRequest("POST", "/user/me/username", updateUsernameData, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accData presenter.Success[identity.UserView]
		err = json.NewDecoder(resp.Body).Decode(&accData)
		require.NoError(t, err)

		assert.Equal(t, updateUsernameData["value"], accData.Data.Username)

		t.Run("should retrieve updated username", func(t *testing.T) {
			resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, token)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var accData presenter.Success[identity.UserView]
			err = json.NewDecoder(resp.Body).Decode(&accData)
			require.NoError(t, err)

			assert.Equal(t, updateUsernameData["value"], accData.Data.Username)

			// Verify that email remains unchanged
			assert.Equal(t, userData.Email, accData.Data.Email.String())
		})
	})

}

func TestPasswordChange(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	userData := ts.newUser(t)
	token := ts.login(t, userData.LoginPayload())

	newPassword := "NewTestPassword123!"

	t.Run("should handle password change flow", func(t *testing.T) {
		// Request password reset (assuming endpoint exists)
		changePassword := map[string]interface{}{
			"oldPassword": userData.Password,
			"newPassword": newPassword,
		}

		resp, err := ts.AuthenticatedRequest("POST", "/user/me/password", changePassword, token)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		t.Run("should unauthorize login with old password", func(t *testing.T) {
			// Attempt login with old password
			resp, err := ts.Request("POST", "/auth/login", userData.LoginPayload(), nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("should login with new password", func(t *testing.T) {
			// Attempt login with new password
			factory := NewTestDataFactory()
			loginData := factory.CreateLoginData(userData.Email, newPassword)
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

	userData := ts.newUser(t)
	newPassword := "NewPassword123!"

	t.Run("should handle password reset flow", func(t *testing.T) {
		// Request password reset
		resetRequest := map[string]interface{}{
			"entry": userData.Email,
		}

		resp, err := ts.Request("POST", "/user/request-password", resetRequest, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		t.Run("should reset password", func(t *testing.T) {
			// Query database to get verification code
			dao := identrepo.NewFactory().NewDAO(context.Background(), ts.server.GetBasePool().Connection())
			user, err := dao.GetUserByEntry(shared.Entry(userData.Email))
			require.NoError(t, err)
			require.NotNil(t, user)

			resetData := map[string]interface{}{
				"code":        user.VerificationCodes[identity.VerificationPasswordReset],
				"newPassword": newPassword,
			}

			resp, err := ts.Request("POST", "/user/"+user.ID.String()+"/password", resetData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			factory := NewTestDataFactory()

			t.Run("should login with new password", func(t *testing.T) {
				loginData := factory.CreateLoginData(userData.Email, newPassword)
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
				loginData := factory.CreateLoginData(userData.Email, userData.Password)
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
		userData := factory.RandomFullUser()

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
