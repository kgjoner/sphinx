package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserData struct {
	ID       uuid.UUID
	Email    string
	Password string
}

func (u UserData) LoginPayload() map[string]interface{} {
	return map[string]interface{}{
		"entry":    u.Email,
		"password": u.Password,
	}
}

// newUser creates a new user and returns the user data
func (ts *TestSuite) newUser(t *testing.T) UserData {
	factory := NewTestDataFactory()
	userData := factory.RandomUser()

	resp, err := ts.Request("POST", "/user", userData, nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var respData httpserver.SuccessResponse[identity.UserLeanView]
	err = json.NewDecoder(resp.Body).Decode(&respData)
	require.NoError(t, err)
	
	return UserData{
		ID:       respData.Data.ID,
		Email:    userData["email"].(string),
		Password: userData["password"].(string),
	}
}

// login performs a login request and returns the access token
func (ts *TestSuite) login(t *testing.T, payload map[string]any) string {
	// Login
	loginResp, err := ts.Request("POST", "/auth/login", payload, nil)
	require.NoError(t, err)
	defer loginResp.Body.Close()

	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	token := loginResult["data"].(map[string]interface{})["accessToken"].(string)
	return token
}
