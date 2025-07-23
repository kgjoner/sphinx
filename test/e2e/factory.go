package e2e

import "github.com/kgjoner/sphinx/test/mocks"

// Test data factory
type TestDataFactory struct{}

func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{}
}

func (f *TestDataFactory) CreateAccountData() map[string]interface{} {
	return map[string]interface{}{
		"email":    "test@example.com",
		"password": "TestPassword123!",
		"username": "testuser",
		"phone":    "+5511999999999",
		"document": "cpf:02496946031",
		"name":     "Test",
		"surname":  "User",
		"address": map[string]interface{}{
			"line1":   "Test Street",
			"number":  "123",
			"city":    "Test City",
			"state":   "SP",
			"country": "Brazil",
			"zipCode": "12345-678",
		},
	}
}

func (f *TestDataFactory) CreateLoginData(email, password string) map[string]interface{} {
	return map[string]interface{}{
		"entry":    email,
		"password": password,
	}
}

func (f *TestDataFactory) SimpleUserLoginData() map[string]interface{} {
	return map[string]interface{}{
		"entry":    mocks.SimpleUserAccount.Email.String(),
		"password": mocks.SimpleUserPassword,
	}
}

func (f *TestDataFactory) CreateApplicationData() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Test App",
		"description": "Test Application",
		"website":     "https://test.com",
	}
}
