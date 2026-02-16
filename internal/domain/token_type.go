package domain

type TokenType string

const (
	TokenTypeVerifyEmail   TokenType = "verify_email"
	TokenTypePasswordReset TokenType = "password_reset"
	TokenTypeMagicLink     TokenType = "magic_link"
)

func (t TokenType) String() string {
	return string(t)
}

func (t TokenType) IsValid() bool {
	return t == TokenTypeVerifyEmail || t == TokenTypePasswordReset || t == TokenTypeMagicLink
}
