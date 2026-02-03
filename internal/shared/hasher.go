package shared

type PasswordHasher interface {
	HashPassword(password string) string
	DoesPasswordMatch(hashed string, password string) bool
}

type DataHasher interface {
	HashData(data string) string
	DoesDataMatch(hashedData, data string) bool
}