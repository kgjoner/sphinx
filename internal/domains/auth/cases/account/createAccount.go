package accountcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type CreateAccount struct {
	AuthRepo authcase.AuthRepo
}

type CreateAccountInput struct {
	auth.AccountCreationFields
	Application auth.Application
}

func (i CreateAccount) Execute(input CreateAccountInput) (*auth.Account, error) {
	acc, err := auth.NewAccount(&input.AccountCreationFields)
	if err != nil {
		return nil, err
	}
	
	err = acc.LinkTo(input.Application)
	if err != nil {
		return nil, err
	}
	
	// TODO: Send emails
	
	_, err = i.AuthRepo.InsertAccount(*acc)
	if err != nil {
		return nil, err
	}
	
	err = i.AuthRepo.UpsertLinks(acc.LinksToPersist()...)
	if err != nil {
		return nil, err
	}
	
	return acc, nil
}