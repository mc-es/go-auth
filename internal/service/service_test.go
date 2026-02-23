package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/internal/apperror"
	"go-auth/internal/domain"
	"go-auth/internal/service"
)

const (
	testAccessTTL  = 15 * time.Minute
	testRefreshTTL = 7 * 24 * time.Hour
)

type mockUserRepo struct {
	saveErr             error
	getByIDUser         *domain.User
	getByIDErr          error
	getByUsernameUser   *domain.User
	getByUsernameErr    error
	getByEmailUser      *domain.User
	getByEmailErr       error
	existsByUsername    bool
	existsByUsernameErr error
	existsByEmail       bool
	existsByEmailErr    error
	savedUser           *domain.User
}

func (m *mockUserRepo) Save(ctx context.Context, user *domain.User) error {
	m.savedUser = user

	return m.saveErr
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return m.getByIDUser, m.getByIDErr
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, u domain.Username) (*domain.User, error) {
	return m.getByUsernameUser, m.getByUsernameErr
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, e domain.Email) (*domain.User, error) {
	return m.getByEmailUser, m.getByEmailErr
}
func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *mockUserRepo) ExistsByUsername(ctx context.Context, u domain.Username) (bool, error) {
	return m.existsByUsername, m.existsByUsernameErr
}

func (m *mockUserRepo) ExistsByEmail(ctx context.Context, e domain.Email) (bool, error) {
	return m.existsByEmail, m.existsByEmailErr
}

type mockSessionRepo struct {
	saveErr       error
	getByToken    *domain.Session
	getByTokenErr error
	updateErr     error
}

func (m *mockSessionRepo) Save(ctx context.Context, session *domain.Session) error { return m.saveErr }
func (m *mockSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	if m.getByTokenErr != nil {
		return nil, m.getByTokenErr
	}

	return m.getByToken, nil
}

func (m *mockSessionRepo) Update(ctx context.Context, session *domain.Session) error {
	return m.updateErr
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error             { return nil }
func (m *mockSessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error { return nil }

type mockPasswordHasher struct {
	hashErr   error
	compareOk bool
}

func (m *mockPasswordHasher) Hash(password string) (domain.Password, error) {
	if m.hashErr != nil {
		return domain.Password{}, m.hashErr
	}

	p, _ := domain.NewPasswordFromHash("stub-hash")

	return p, nil
}
func (m *mockPasswordHasher) Compare(plain string, hash domain.Password) bool { return m.compareOk }

type mockOpaqueTokenManager struct {
	generateToken string
	generateErr   error
	hashResult    string
	hashErr       error
}

func (m *mockOpaqueTokenManager) Generate() (string, error) {
	if m.generateErr != nil {
		return "", m.generateErr
	}

	if m.generateToken != "" {
		return m.generateToken, nil
	}

	return "refresh-token", nil
}

func (m *mockOpaqueTokenManager) Hash(token string) (string, error) {
	if m.hashErr != nil {
		return "", m.hashErr
	}

	if m.hashResult != "" {
		return m.hashResult, nil
	}

	return "hashed-" + token, nil
}

type mockAccessTokenManager struct {
	generateToken string
	generateErr   error
}

func (m *mockAccessTokenManager) Generate(claims domain.AccessClaims) (string, error) {
	if m.generateErr != nil {
		return "", m.generateErr
	}

	if m.generateToken != "" {
		return m.generateToken, nil
	}

	return "access-token", nil
}

func (m *mockAccessTokenManager) Validate(token string) (*domain.AccessClaims, error) {
	return nil, nil
}

func newTestService(
	ur domain.UserRepository,
	sr domain.SessionRepository,
	ph domain.PasswordHasher,
	ot domain.OpaqueTokenManager,
	at domain.AccessTokenManager,
) (service.Service, error) {
	return service.NewService(&service.Config{
		UserRepo:           ur,
		SessionRepo:        sr,
		PasswordHasher:     ph,
		OpaqueTokenManager: ot,
		AccessTokenManager: at,
		AccessTokenTTL:     testAccessTTL,
		RefreshTokenTTL:    testRefreshTTL,
	})
}

// testDeps holds optional test doubles; nil fields are replaced with fresh no-op mocks.
type testDeps struct {
	UserRepo    *mockUserRepo
	SessionRepo *mockSessionRepo
	Hasher      *mockPasswordHasher
	Opaque      *mockOpaqueTokenManager
	Access      *mockAccessTokenManager
}

// newTestServiceWith builds a service from d; any nil dep is filled with a default no-op mock.
func newTestServiceWith(d testDeps) (service.Service, error) {
	if d.UserRepo == nil {
		d.UserRepo = &mockUserRepo{}
	}

	if d.SessionRepo == nil {
		d.SessionRepo = &mockSessionRepo{}
	}

	if d.Hasher == nil {
		d.Hasher = &mockPasswordHasher{}
	}

	if d.Opaque == nil {
		d.Opaque = &mockOpaqueTokenManager{}
	}

	if d.Access == nil {
		d.Access = &mockAccessTokenManager{}
	}

	return newTestService(d.UserRepo, d.SessionRepo, d.Hasher, d.Opaque, d.Access)
}

func mustVerifiedUser(t *testing.T, username, email, passHash string) *domain.User {
	t.Helper()

	un, _ := domain.NewUsername(username)
	em, _ := domain.NewEmail(email)
	pass, _ := domain.NewPasswordFromHash(passHash)
	u, err := domain.NewUser(un, em, pass, "First", "Last")
	require.NoError(t, err)
	require.NoError(t, u.Verify())

	return u
}

var farFutureExpiry = time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

func mustSession(t *testing.T, userID uuid.UUID, expiresIn time.Duration, revoked bool) *domain.Session {
	t.Helper()

	if expiresIn <= 0 {
		now := time.Now().UTC()

		return &domain.Session{
			ID:        uuid.New(),
			UserID:    userID,
			Token:     "token-hash",
			UserAgent: "ua",
			ClientIP:  "1.2.3.4",
			ExpiresAt: now.Add(-time.Hour),
			RevokedAt: nil,
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now,
		}
	}

	s, err := domain.NewSession(userID, "token-hash", "ua", "1.2.3.4", farFutureExpiry)
	require.NoError(t, err)

	if revoked {
		require.NoError(t, s.Revoke())
	}

	return s
}

func assertAppErrorCode(t *testing.T, err error, code apperror.Code) {
	t.Helper()
	require.Error(t, err)

	var ae *apperror.Error
	require.True(t, errors.As(err, &ae), "expected *apperror.Error, got %T", err)
	assert.Equal(t, code, ae.Code)
}

func TestNewService(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		svc, err := newTestServiceWith(testDeps{})
		require.NoError(t, err)
		require.NotNil(t, svc)
	})

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()

		_, err := service.NewService(nil)
		require.Error(t, err)
	})

	t.Run("missing user repo", func(t *testing.T) {
		t.Parallel()

		_, err := service.NewService(&service.Config{
			SessionRepo:        &mockSessionRepo{},
			PasswordHasher:     &mockPasswordHasher{},
			OpaqueTokenManager: &mockOpaqueTokenManager{},
			AccessTokenManager: &mockAccessTokenManager{},
			AccessTokenTTL:     testAccessTTL,
			RefreshTokenTTL:    testRefreshTTL,
		})
		require.Error(t, err)
	})
}
