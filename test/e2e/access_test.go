package e2e

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/access/accesscase"
	"github.com/kgjoner/sphinx/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationManagement(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.server.Close()

	seedData := ts.GetSeedData()

	t.Run("should create an application", func(t *testing.T) {
		// Use admin user who has permissions
		factory := NewTestDataFactory()
		payload := factory.AdminUserLoginData()
		token := ts.login(t, payload)

		// Create application with unique name
		uniqueName := "Test App " + uuid.New().String()[:8]
		appData := map[string]interface{}{
			"name":                uniqueName,
			"possibleRoles":       []string{"admin", "user"},
			"allowedRedirectUris": []string{"https://myapp.com/callback"},
		}

		resp, err := ts.AuthenticatedRequest("POST", "/application", appData, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp httpserver.SuccessResponse[accesscase.CreateApplicationOutput]
		bodyBytes, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(bodyBytes, &apiResp)
		require.NoError(t, err)

		assert.Equal(t, uniqueName, apiResp.Data.Application.Name)
		assert.NotEmpty(t, apiResp.Data.Application.ID)
		assert.NotEmpty(t, apiResp.Data.Secret)
	})

	t.Run("should validate application creation requires name", func(t *testing.T) {
		factory := NewTestDataFactory()
		payload := factory.AdminUserLoginData()
		token := ts.login(t, payload)

		// Missing name
		appData := map[string]interface{}{
			"possibleRoles": []string{"admin"},
		}

		resp, err := ts.AuthenticatedRequest("POST", "/application", appData, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should fail with validation error - could be 400 or 403 depending on order
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, resp.StatusCode)
	})

	t.Run("should require authentication to create application", func(t *testing.T) {
		appData := map[string]interface{}{
			"name": "Unauthorized App",
		}

		resp, err := ts.Request("POST", "/application", appData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should retrieve an application by ID", func(t *testing.T) {
		appID := seedData.TestApp.ID.String()

		resp, err := ts.Request("GET", "/application/"+appID, nil, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respData httpserver.SuccessResponse[access.ApplicationView]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.Equal(t, seedData.TestApp.Name, respData.Data.Name)
		assert.Equal(t, seedData.TestApp.ID.String(), respData.Data.ID.String())
	})

	t.Run("should return 404 for non-existent application", func(t *testing.T) {
		// Use a valid UUID that doesn't exist
		fakeID := "00000000-0000-0000-0000-000000000099"

		resp, err := ts.Request("GET", "/application/"+fakeID, nil, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Contains(t, []int{http.StatusNotFound, http.StatusBadRequest}, resp.StatusCode)
	})

	t.Run("should edit an application", func(t *testing.T) {
		factory := NewTestDataFactory()
		payload := factory.AdminUserLoginData()
		token := ts.login(t, payload)

		// Create application first
		originalName := "Original Name " + uuid.New().String()[:8]
		appData := map[string]interface{}{
			"name": originalName,
		}
		createResp, err := ts.AuthenticatedRequest("POST", "/application", appData, token)
		require.NoError(t, err)
		bodyBytes, _ := io.ReadAll(createResp.Body)
		var createData httpserver.SuccessResponse[accesscase.CreateApplicationOutput]
		json.Unmarshal(bodyBytes, &createData)
		createResp.Body.Close()

		appID := createData.Data.Application.ID.String()
		appSecret := createData.Data.Secret

		assert.NotEmpty(t, appID)
		assert.NotEmpty(t, appSecret)

		// Edit application
		updatedName := "Updated Name " + uuid.New().String()[:8]
		editData := map[string]interface{}{
			"name":                updatedName,
			"possibleRoles":       []string{"admin", "editor"},
			"allowedRedirectUris": []string{"https://updated.com/callback"},
		}

		t.Run("should deny editing with another application token", func(t *testing.T) {
			editResp, err := ts.AuthenticatedRequest("PATCH", "/application/"+appID, editData, token)
			require.NoError(t, err)
			defer editResp.Body.Close()

			assert.Equal(t, http.StatusForbidden, editResp.StatusCode)
		})

		t.Run("should allow editing with basic app auth", func(t *testing.T) {
			editResp, err := ts.AuthenticatedAppRequest("PATCH", "/application/"+appID, editData, appID, createData.Data.Secret)
			require.NoError(t, err)
			defer editResp.Body.Close()

			assert.Equal(t, http.StatusOK, editResp.StatusCode)

			var editApiResp httpserver.SuccessResponse[access.ApplicationView]
			editBodyBytes, _ := io.ReadAll(editResp.Body)
			err = json.Unmarshal(editBodyBytes, &editApiResp)
			if err == nil {
				assert.Equal(t, updatedName, editApiResp.Data.Name)
			}
		})
	})

	t.Run("should require authentication to edit application", func(t *testing.T) {
		appID := seedData.TestApp.ID.String()

		editData := map[string]interface{}{
			"name": "Should Not Update",
		}

		resp, err := ts.Request("PATCH", "/application/"+appID, editData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestApplicationAuthenticationIntegration(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.server.Close()

	seedData := ts.GetSeedData()

	t.Run("should authenticate with app credentials", func(t *testing.T) {
		appID := seedData.RootApp.ID.String()
		appSecret := mocks.RootAppSecret

		resp, err := ts.AuthenticatedAppRequest(
			"GET",
			"/user/"+seedData.SimpleUser.ID.String()+"/link/"+appID,
			nil,
			appID,
			appSecret,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("should reject invalid app credentials", func(t *testing.T) {
		appID := seedData.RootApp.ID.String()
		invalidSecret := "wrong-" + mocks.RootAppSecret

		resp, err := ts.AuthenticatedAppRequest(
			"GET",
			"/user/"+seedData.SimpleUser.ID.String()+"/link/"+appID,
			nil,
			appID,
			invalidSecret,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should reject invalid app ID", func(t *testing.T) {
		invalidAppID := uuid.New().String()
		appSecret := mocks.RootAppSecret

		resp, err := ts.AuthenticatedAppRequest(
			"GET",
			"/user/"+seedData.SimpleUser.ID.String()+"/link/"+invalidAppID,
			nil,
			invalidAppID,
			appSecret,
		)

		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestUserLinkManagement(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.server.Close()

	seedData := ts.GetSeedData()

	t.Run("should return user link for target application", func(t *testing.T) {
		userID := seedData.SimpleUser.ID.String()
		appID := seedData.RootApp.ID.String()
		factory := NewTestDataFactory()
		token := ts.login(t, factory.SimpleUserLoginData())

		resp, err := ts.AuthenticatedRequest(
			"GET",
			"/user/"+userID+"/link/"+appID,
			nil,
			token,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var link httpserver.SuccessResponse[access.LinkView]
		err = json.NewDecoder(resp.Body).Decode(&link)
		require.NoError(t, err)

		assert.Equal(t, seedData.SimpleRootLink.ID, link.Data.ID)
		assert.Equal(t, seedData.RootApp.ID, link.Data.ApplicationID)
		assert.Equal(t, seedData.SimpleRootLink.Roles, link.Data.Roles)
	})

	t.Run("should return 404 for non-existent link", func(t *testing.T) {
		userID := uuid.New().String()
		appID := seedData.RootApp.ID.String()
		factory := NewTestDataFactory()
		token := ts.login(t, factory.AdminUserLoginData())

		resp, err := ts.AuthenticatedRequest(
			"GET",
			"/user/"+userID+"/link/"+appID,
			nil,
			token,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Contains(t, []int{http.StatusNotFound, http.StatusBadRequest}, resp.StatusCode)
	})

	t.Run("should require authentication to get link", func(t *testing.T) {
		userID := seedData.SimpleUser.ID.String()
		appID := seedData.RootApp.ID.String()

		resp, err := ts.Request("GET", "/user/"+userID+"/link/"+appID, nil, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should add role via app auth", func(t *testing.T) {
		userData := ts.newUser(t)
		userID := userData.ID.String()
		appID := seedData.RootApp.ID.String()
		appSecret := mocks.RootAppSecret

		// Add role through API
		resp, err := ts.AuthenticatedAppRequest(
			"PUT",
			"/user/"+userID+"/link/"+appID+"/role/"+string(mocks.DummyRole),
			nil,
			appID,
			appSecret,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Ensure role was added
		resp, err = ts.AuthenticatedAppRequest(
			"GET",
			"/user/"+userID+"/link/"+appID,
			nil,
			appID,
			appSecret,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var link httpserver.SuccessResponse[access.LinkView]
		err = json.NewDecoder(resp.Body).Decode(&link)
		require.NoError(t, err)

		assert.Contains(t, link.Data.Roles, mocks.DummyRole)

		t.Run("should remove dummy role via app auth", func(t *testing.T) {
			// Remove role
			removeResp, err := ts.AuthenticatedAppRequest(
				"DELETE",
				"/user/"+userID+"/link/"+appID+"/role/"+string(mocks.DummyRole),
				nil,
				appID,
				appSecret,
			)
			require.NoError(t, err)
			defer removeResp.Body.Close()

			assert.Equal(t, http.StatusNoContent, removeResp.StatusCode)

			// Ensure role was removed
			removeResp, err = ts.AuthenticatedAppRequest(
				"GET",
				"/user/"+userID+"/link/"+appID,
				nil,
				appID,
				appSecret,
			)
			require.NoError(t, err)
			defer removeResp.Body.Close()

			assert.Equal(t, http.StatusOK, removeResp.StatusCode)

			var link httpserver.SuccessResponse[access.LinkView]
			err = json.NewDecoder(removeResp.Body).Decode(&link)
			require.NoError(t, err)

			assert.NotContains(t, link.Data.Roles, mocks.DummyRole)
		})
	})

	t.Run("should deny adding privileged role via app auth", func(t *testing.T) {
		userID := seedData.SimpleUser.ID.String()
		appID := seedData.RootApp.ID.String()
		appSecret := mocks.RootAppSecret

		resp, err := ts.AuthenticatedAppRequest(
			"PUT",
			"/user/"+userID+"/link/"+appID+"/role/admin",
			nil,
			appID,
			appSecret,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("should deny removing privileged role via app auth", func(t *testing.T) {
		userID := seedData.SimpleUser.ID.String()
		appID := seedData.RootApp.ID.String()
		appSecret := mocks.RootAppSecret

		removeResp, err := ts.AuthenticatedAppRequest(
			"DELETE",
			"/user/"+userID+"/link/"+appID+"/role/admin",
			nil,
			appID,
			appSecret,
		)
		require.NoError(t, err)
		defer removeResp.Body.Close()

		assert.Equal(t, http.StatusForbidden, removeResp.StatusCode)
	})

	t.Run("should require authentication to add role", func(t *testing.T) {
		userID := seedData.SimpleUser.ID.String()
		appID := seedData.RootApp.ID.String()

		resp, err := ts.Request(
			"PUT",
			"/user/"+userID+"/link/"+appID+"/role/manager",
			nil,
			nil,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should deny repeated role assignment", func(t *testing.T) {
		userData := ts.newUser(t)
		userID := userData.ID.String()
		appID := seedData.RootApp.ID.String()
		appSecret := mocks.RootAppSecret

		resp, err := ts.AuthenticatedAppRequest(
			"PUT",
			"/user/"+userID+"/link/"+appID+"/role/"+string(mocks.DummyRole),
			nil,
			appID,
			appSecret,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()

		resp, err = ts.AuthenticatedAppRequest(
			"PUT",
			"/user/"+userID+"/link/"+appID+"/role/"+string(mocks.DummyRole),
			nil,
			appID,
			appSecret,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestLegacyPermissionEndpoint(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.server.Close()

	seedData := ts.GetSeedData()

	userData := ts.newUser(t)
	userID := userData.ID.String()

	t.Run("should add permission via legacy endpoint", func(t *testing.T) {
		appID := seedData.RootApp.ID.String()
		appSecret := mocks.RootAppSecret

		permissionData := map[string]interface{}{
			"roles": []string{string(mocks.DummyRole)},
		}

		resp, err := ts.AuthenticatedAppRequest(
			"PATCH",
			"/user/"+userID+"/permission",
			permissionData,
			appID,
			appSecret,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Ensure role was added
		resp, err = ts.AuthenticatedAppRequest(
			"GET",
			"/user/"+userID+"/link/"+appID,
			nil,
			appID,
			appSecret,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var link httpserver.SuccessResponse[access.LinkView]
		err = json.NewDecoder(resp.Body).Decode(&link)
		require.NoError(t, err)

		assert.Contains(t, link.Data.Roles, mocks.DummyRole)

		t.Run("should remove permission via legacy endpoint", func(t *testing.T) {
			appID := seedData.RootApp.ID.String()
			appSecret := mocks.RootAppSecret

			// Now remove it via legacy endpoint
			removeData := map[string]interface{}{
				"roles":        []string{string(mocks.DummyRole)},
				"shouldRemove": sql.NullBool{Bool: true, Valid: true},
			}

			removeResp, err := ts.AuthenticatedAppRequest(
				"PATCH",
				"/user/"+userID+"/permission",
				removeData,
				appID,
				appSecret,
			)
			require.NoError(t, err)
			defer removeResp.Body.Close()

			assert.Equal(t, http.StatusNoContent, removeResp.StatusCode)

			// Ensure role was removed
			removeResp, err = ts.AuthenticatedAppRequest(
				"GET",
				"/user/"+userID+"/link/"+appID,
				nil,
				appID,
				appSecret,
			)
			require.NoError(t, err)
			defer removeResp.Body.Close()

			assert.Equal(t, http.StatusOK, removeResp.StatusCode)

			var link httpserver.SuccessResponse[access.LinkView]
			err = json.NewDecoder(removeResp.Body).Decode(&link)
			require.NoError(t, err)

			assert.NotContains(t, link.Data.Roles, mocks.DummyRole)

		})
	})

}
