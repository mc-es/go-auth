package domain

type PasswordHasher interface {
	Hash(password string) (Password, error)
	Compare(plainText string, hash Password) bool
}
