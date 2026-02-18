package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go-auth/internal/domain"
)

type jwtManager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type jwtClaims struct {
	jwt.RegisteredClaims

	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
}

func NewJWT(secret, issuer string, accessTTL, refreshTTL time.Duration) (domain.TokenManager, error) {
	if secret == "" {
		return nil, domain.ErrTokenSecretRequired
	}

	if accessTTL <= 0 {
		return nil, domain.ErrTokenAccessTTLRequired
	}

	if refreshTTL <= 0 {
		return nil, domain.ErrTokenRefreshTTLRequired
	}

	return &jwtManager{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

func (m *jwtManager) Generate(claims domain.Claims) (string, error) {
	now := time.Now().UTC()

	var ttl time.Duration

	switch claims.Type {
	case domain.TokenTypeRefresh:
		ttl = m.refreshTTL
	case domain.TokenTypeAccess, domain.TokenTypeVerifyEmail, domain.TokenTypePasswordReset, domain.TokenTypeMagicLink:
		ttl = m.accessTTL
	default:
		return "", domain.ErrTokenTypeInvalid
	}

	exp := now.Add(ttl)

	rc := jwt.RegisteredClaims{
		Issuer:    m.issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
		NotBefore: jwt.NewNumericDate(now),
	}
	claimsData := jwtClaims{
		RegisteredClaims: rc,
		UserID:           claims.UserID.String(),
		Role:             claims.Role.String(),
		TokenType:        claims.Type.String(),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsData)

	signed, err := tok.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (m *jwtManager) Validate(token string) (*domain.Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwtClaims{}, m.keyFunc, jwt.WithIssuer(m.issuer))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}

		return nil, domain.ErrTokenInvalid
	}

	claims, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return nil, domain.ErrTokenInvalid
	}

	return m.parseClaims(claims)
}

func (m *jwtManager) keyFunc(t *jwt.Token) (any, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, domain.ErrTokenInvalid
	}

	return m.secret, nil
}

func (m *jwtManager) parseClaims(claims *jwtClaims) (*domain.Claims, error) {
	tokenType := domain.TokenType(claims.TokenType)
	if !tokenType.IsValid() {
		return nil, domain.ErrTokenTypeInvalid
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	role, err := domain.NewRole(claims.Role)
	if err != nil {
		return nil, domain.ErrRoleInvalid
	}

	var exp time.Time
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}

	return &domain.Claims{
		UserID:    userID,
		Role:      role,
		Type:      tokenType,
		ExpiresAt: exp,
	}, nil
}
