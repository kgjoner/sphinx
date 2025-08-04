package e2e

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/test/mocks"
)

// Test data factory
type TestDataFactory struct{}

func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{}
}

// Return a prefilled user with all fields
func (f *TestDataFactory) FullUser() map[string]interface{} {
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

func (f *TestDataFactory) RandomUser() map[string]interface{} {
	return map[string]interface{}{
		"email":    GenerateEmail(),
		"password": GenerateStrongPassword(),
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
		"entry":    mocks.SimpleUserUser.Email.String(),
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

func GenerateEmail() string {
	return "test-" + uuid.New().String()[:8] + "@example.com"
}

func GenerateUsername() string {
	return "user-" + uuid.New().String()[:8]
}

func GenerateStrongPassword() string {
	return "StrongPass123!@#" + uuid.New().String()[:4]
}

// Generate a valid phone number for Brazil.
//
// You may pass a format in the form +X(X)XXXX-XXXX. Requires 10 or 11 X's. The first 2 X's will be replaced with the country code and area code.
// If no format is provided, or if invalid, it will return a valid PhoneNumber as +5511XXXXXXXX.
func GeneratePhone(format ...string) string {
	if len(format) > 0 {
		xCount := strings.Count(format[0], "X")
		if xCount == 10 || xCount == 11 {
			result := strings.Replace(format[0], "X", "55", 1)
			result = strings.Replace(result, "X", "11", 1)
			result = strings.ReplaceAll(result, "X", fmt.Sprintf("%d", rand.Intn(10)))
			return result
		}
	}

	return fmt.Sprintf("+5511%d%04d%04d",
		9,                    // Mobile prefix
		1000+rand.Intn(8999), // First 4 digits (1000-9999)
		rand.Intn(10000))     // Last 4 digits (0000-9999)
}

// Generate a valid CPF (Brazilian individual taxpayer ID) to be used as Document.
//
// You may pass a format in the form XXX.XXX.XXX-XX. Requires 11 X's.
// If no format is provided, or if invalid, it will return a valid CPF without formatting (only numbers).
func GenerateCPF(format ...string) string {
	digits := make([]int, 11)

	// Generate first 9 digits randomly
	for i := 0; i < 9; i++ {
		digits[i] = rand.Intn(10)
	}

	// Ensure they are not all the same digit
	for i := 1; i < 9; i++ {
		if digits[i] != digits[0] {
			break
		}
		if i == 8 {
			// If all digits are the same, regenerate the first 9 digits
			return GenerateCPF(format...)
		}
	}

	// Calculate first check digit
	sum := 0
	for i := 0; i < 9; i++ {
		sum += digits[i] * (10 - i)
	}
	remainder := sum % 11
	if remainder < 2 {
		digits[9] = 0
	} else {
		digits[9] = 11 - remainder
	}

	// Calculate second check digit
	sum = 0
	for i := 0; i < 10; i++ {
		sum += digits[i] * (11 - i)
	}
	remainder = sum % 11
	if remainder < 2 {
		digits[10] = 0
	} else {
		digits[10] = 11 - remainder
	}

	skeleton := "%d%d%d%d%d%d%d%d%d%d%d"
	if len(format) > 0 {
		xCount := strings.Count(format[0], "X")
		if xCount == 11 {
			skeleton = strings.ReplaceAll(format[0], "X", "%d")
		}
	}

	return fmt.Sprintf(skeleton,
		digits[0], digits[1], digits[2], digits[3], digits[4],
		digits[5], digits[6], digits[7], digits[8], digits[9], digits[10])
}
