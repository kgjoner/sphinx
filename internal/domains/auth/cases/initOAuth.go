package authcase

import (
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/config/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type InitOAuth struct {
	AuthRepo AuthRepo
}

type InitOAuthInput struct {
	Entry       string
	Password    string
	Application auth.Application `json:"-"`
}

func (i InitOAuth) Execute(input InitOAuthInput) (*string, error) {
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

	return &code, nil
}
