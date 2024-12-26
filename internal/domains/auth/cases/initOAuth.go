package authcase

import (
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type InitOAuth struct {
	AuthRepo AuthRepo
}

type InitOAuthInput struct {
	Entry       string           `validate:"required"`
	Password    string           `validate:"required"`
	State       string           `validate:"required"`
	Application auth.Application `json:"-" validate:"required"`
}

func (i InitOAuth) Execute(input InitOAuthInput) (*InitOAuthOutput, error) {
	err := validator.Validate(input)
	if err != nil {
		return nil, err
	}

	acc, err := i.AuthRepo.GetAccountByEntry(input.Entry)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, normalizederr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
	}

	err = acc.AuthenticateViaPassword(input.Password)
	if err != nil {
		return nil, err
	}

	code, err := acc.InitOAuth(input.Application)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertLinks(acc.LinksToPersist()...)
	if err != nil {
		return nil, err
	}

	return &InitOAuthOutput{
		code,
		input.State,
	}, nil
}

type InitOAuthOutput struct {
	Code  string `json:"code"`
	State string `json:"state"`
}
