package oauthcase

import (
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type IssueGrant struct {
	AuthRepo authcase.AuthRepo
}

type IssueGrantInput struct {
	Entry               string           `validate:"required"`
	Password            string           `validate:"required"`
	CodeChallenge       string           `json:"code_challenge"`
	CodeChallengeMethod string           `json:"code_challenge_method"`
	Application         auth.Application `json:"-" validate:"required"`
}

func (i IssueGrant) Execute(input IssueGrantInput) (*string, error) {
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

	code, err := acc.InitOAuth(input.Application, input.CodeChallenge, input.CodeChallengeMethod)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertLinks(acc.LinksToPersist()...)
	if err != nil {
		return nil, err
	}

	return &code, nil
}
