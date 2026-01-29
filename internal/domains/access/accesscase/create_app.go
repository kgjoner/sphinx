package accesscase

import (
	"github.com/kgjoner/cornucopia/v2/utils/pwdgen"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/shared"
)

type CreateApplication struct {
	AccessRepo access.Repo
	Hasher     shared.PasswordHasher
}

type CreateApplicationInput struct {
	access.ApplicationCreationFields
	Actor shared.Actor `json:"-"`
}

func (i CreateApplication) Execute(input CreateApplicationInput) (out CreateApplicationOutput, err error) {
	if err := access.CanCreateApplication(&input.Actor); err != nil {
		return out, err
	}

	// TODO: remove direct reference to pwdgen package
	// and use shared.PasswordGenerator interface instead
	secret := pwdgen.GeneratePassword(42, "lower", "upper", "number")

	hashPw, err := shared.NewHashedPassword(secret, i.Hasher)
	if err != nil {
		return out, err
	}

	input.ApplicationCreationFields.Secret = *hashPw
	app, err := access.NewApplication(&input.ApplicationCreationFields)
	if err != nil {
		return out, err
	}

	err = i.AccessRepo.InsertApplication(app)
	if err != nil {
		return out, err
	}

	return CreateApplicationOutput{
		Application: app.View(),
		Secret:      secret,
	}, nil
}

type CreateApplicationOutput struct {
	Application access.ApplicationView `json:"application"`
	Secret      string                 `json:"secret"`
}
