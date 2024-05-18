package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/config/errcode"
	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	config.Must()

	acc := &Account{
		Id: uuid.New(),
	}
	sId := uuid.New()
	acc.ActiveSessions = []Session{
		{
			Id:       sId,
			IsActive: true,
			Application: Application{
				Id: uuid.New(),
			},
		},
	}

	token, err := newAuthToken(authTokenCreationFields{
		*acc,
		sId,
		false,
	})

	assert.Nil(t, err)

	parsedToken, err := ParseAuthToken(token.String())
	assert.Nil(t, err)
	assert.NotEmpty(t, parsedToken)

	token.Claims.Exp = time.Now().Add(-2 * time.Second).Unix()
	modifiedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token.Claims)
	modifiedTokenStr, err := modifiedToken.SignedString([]byte(config.Env.JWT.SECRET))
	assert.Nil(t, err)
	_, err = ParseAuthToken(modifiedTokenStr)
	normErr := err.(normalizederr.NormalizedError)
	assert.Equal(t, normErr.Code, errcode.ExpiredAccess)
}

func TestRefreshToken(t *testing.T) {
	config.Must()

	acc := &Account{
		Id: uuid.New(),
	}
	sId := uuid.New()
	acc.ActiveSessions = []Session{
		{
			Id:       sId,
			IsActive: true,
			Application: Application{
				Id: uuid.New(),
			},
		},
	}

	token, err := newAuthToken(authTokenCreationFields{
		*acc,
		sId,
		true,
	})

	assert.Nil(t, err)

	parsedToken, err := ParseAuthToken(token.String())
	assert.Nil(t, err)
	assert.NotEmpty(t, parsedToken)
	assert.True(t, parsedToken.IsRefresh())

	token.Claims.Iat = time.Now().Add(-2 * time.Second).Add(-1 * time.Second * time.Duration(config.Env.JWT.REFRESH_LIFETIME_IN_SEC)).Unix()
	token.Claims.Exp = time.Now().Add(-2 * time.Second).Unix()
	modifiedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token.Claims)
	modifiedTokenStr, err := modifiedToken.SignedString([]byte(config.Env.JWT.SECRET))
	assert.Nil(t, err)
	_, err = ParseAuthToken(modifiedTokenStr)
	normErr := err.(normalizederr.NormalizedError)
	assert.Equal(t, normErr.Code, errcode.ExpiredSession)
}
