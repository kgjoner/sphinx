package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/config"
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
			Id: sId,
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
}
