package accountcase

import authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"

type CheckEntryExistence struct {
	AuthRepo authcase.AuthRepo
}

type CheckEntryExistenceInput struct {
	Entry string
}

func (i CheckEntryExistence) Execute(input CheckEntryExistenceInput) (bool, error) {
	acc, err := i.AuthRepo.GetAccountByEntry(input.Entry);
	if err != nil {
	 return false, err
	}

	res := false
	if acc != nil {
		res = true
	}

	return res, nil
}
