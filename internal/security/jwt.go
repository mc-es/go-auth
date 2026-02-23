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
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

type jwtClaims struct {
	jwt.RegisteredClaims

	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func NewJWT(secret, issuer string, accessTTL time.Duration) (domain.AccessTokenManager, error) {
	if secret == "" {
		return nil, domain.ErrTokenSecretRequired
	}

	if accessTTL <= 0 {
		return nil, domain.ErrTokenAccessTTLRequired
	}

	return &jwtManager{
		secret:    []byte(secret),
		issuer:    issuer,
		accessTTL: accessTTL,
	}, nil
}

func (m *jwtManager) Generate(claims domain.AccessClaims) (string, error) {
	now := time.Now().UTC()

	rc := jwt.RegisteredClaims{
		Issuer:    m.issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		NotBefore: jwt.NewNumericDate(now),
	}
	claimsData := jwtClaims{
		RegisteredClaims: rc,
		UserID:           claims.UserID.String(),
		Role:             claims.Role.String(),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsData)

	signed, err := tok.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (m *jwtManager) Validate(token string) (*domain.AccessClaims, error) {
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

func (m *jwtManager) parseClaims(claims *jwtClaims) (*domain.AccessClaims, error) {
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	role, err := domain.NewRole(claims.Role)
	if err != nil {
		return nil, domain.ErrRoleInvalid
	}

	return &domain.AccessClaims{
		UserID: userID,
		Role:   role,
	}, nil
}
