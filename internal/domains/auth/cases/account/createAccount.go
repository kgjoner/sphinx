package accountcase

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kgjoner/cornucopia/helpers/normalizederr"
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

	id, err := i.AuthRepo.InsertAccount(*acc)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			pattern := regexp.MustCompile("account_(.+)_key")
			matches := pattern.FindStringSubmatch(err.Error())
			msg := fmt.Sprintf("%v has already registered", matches[1])
			return nil, normalizederr.NewRequestError(msg, "")
		}
		return nil, err
	}

	// TODO: Send emails

	acc.InternalId = id
	err = acc.LinkTo(input.Application)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertLinks(acc.LinksToPersist()...)
	if err != nil {
		return nil, err
	}

	return acc, nil
}
