package usercase

import (
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
	"github.com/sirupsen/logrus"
)

type SignUp struct {
	IdentityRepo identity.Repo
	AccessRepo   access.Repo
	Hasher       shared.PasswordHasher
	Mailer       shared.Mailer
}

type SignUpInput struct {
	identity.UserCreationFields
	Password  string
	Languages []string `json:"-"`
}

func (i SignUp) Execute(input SignUpInput) (out identity.UserLeanView, err error) {
	user, err := i.ExecuteEntity(input)
	if err != nil {
		return out, err
	}

	return user.LeanView(), nil
}

// ExecuteEntity is application-internal: returns the entity for chaining.
// Only used by internal application layer (e.g., ExternalAuth).
func (i SignUp) ExecuteEntity(input SignUpInput) (*identity.User, error) {
	hashPw, err := shared.NewHashedPassword(input.Password, i.Hasher)
	if err != nil {
		return nil, err
	}

	input.UserCreationFields.Password = *hashPw
	user, err := identity.NewUser(&input.UserCreationFields, i.Hasher)
	if err != nil {
		return nil, err
	}

	app, err := i.AccessRepo.GetApplicationByID(uuid.MustParse(config.Env.ROOT_APP_ID))
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, access.ErrApplicationNotFound
	}

	err = i.IdentityRepo.InsertUser(user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, identity.ErrDuplicateEntry
		}
		return nil, err
	}

	link, err := app.NewLink(user.ID)
	if err != nil {
		return nil, err
	}

	err = i.AccessRepo.UpsertLinks(*link)
	if err != nil {
		return nil, err
	}

	// Send email
	err = i.Mailer.Send(
		user.Email,
		identity.EmailWelcome{
			UserName: user.Name(),
			UserID:   user.ID.String(),
			Code:     user.VerificationCodes[identity.VerificationEmail],
		},
		input.Languages...,
	)
	if err != nil {
		i.handleError(err, *user)
	}

	return user, nil
}

func (i SignUp) handleError(err error, target identity.User) {
	logrus.WithFields(logrus.Fields{
		"Kind":  "Mail Failed",
		"Path":  "UserCreation",
		"Actor": target.ID,
	}).Log(logrus.ErrorLevel, err.Error())
}
