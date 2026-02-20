package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go-auth/internal/domain"
)

const (
	userTestUsername  = "alice"
	userTestEmail     = "alice@example.com"
	userTestFirstName = "Alice"
	userTestLastName  = "Doe"
	userTestPassHash  = "$argon2id$v=19$m=65536,t=3,p=2$salt$hash"
)

func mustUser(t *testing.T) *domain.User {
	t.Helper()

	username, _ := domain.NewUsername(userTestUsername)
	email, _ := domain.NewEmail(userTestEmail)
	password, _ := domain.NewPasswordFromHash(userTestPassHash)
	u, err := domain.NewUser(username, email, password, userTestFirstName, userTestLastName)
	assert.NoError(t, err)

	return u
}

func TestNewUser(t *testing.T) {
	username, _ := domain.NewUsername(userTestUsername)
	email, _ := domain.NewEmail(userTestEmail)
	password, _ := domain.NewPasswordFromHash(userTestPassHash)

	tests := []struct {
		name      string
		firstName string
		lastName  string
		wantErr   error
	}{
		{
			name:      "valid",
			firstName: userTestFirstName,
			lastName:  userTestLastName,
			wantErr:   nil,
		},
		{
			name:      "trim",
			firstName: "  Alice  ",
			lastName:  "  Doe  ",
			wantErr:   nil,
		},
		{
			name:      "empty first",
			firstName: "",
			lastName:  userTestLastName,
			wantErr:   domain.ErrFirstNameRequired,
		},
		{
			name:      "empty last",
			firstName: userTestFirstName,
			lastName:  "",
			wantErr:   domain.ErrLastNameRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NewUser(username, email, password, tt.firstName, tt.lastName)
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr != nil {
				assert.Nil(t, got)

				return
			}

			assert.NotEqual(t, uuid.Nil, got.ID)
			assert.Equal(t, domain.StatusActivated, got.Status)
			assert.Nil(t, got.VerifiedAt)
			assert.Equal(t, "Alice Doe", got.FullName())
		})
	}
}

func TestNewUserZero(t *testing.T) {
	username, _ := domain.NewUsername(userTestUsername)
	email, _ := domain.NewEmail(userTestEmail)
	password, _ := domain.NewPasswordFromHash(userTestPassHash)

	t.Run("zero username", func(t *testing.T) {
		got, err := domain.NewUser(domain.Username{}, email, password, userTestFirstName, userTestLastName)
		assert.ErrorIs(t, err, domain.ErrUsernameRequired)
		assert.Nil(t, got)
	})

	t.Run("zero email", func(t *testing.T) {
		got, err := domain.NewUser(username, domain.Email{}, password, userTestFirstName, userTestLastName)
		assert.ErrorIs(t, err, domain.ErrEmailRequired)
		assert.Nil(t, got)
	})

	t.Run("zero password", func(t *testing.T) {
		got, err := domain.NewUser(username, email, domain.Password{}, userTestFirstName, userTestLastName)
		assert.ErrorIs(t, err, domain.ErrPasswordRequired)
		assert.Nil(t, got)
	})
}

func TestUserCanLogin(t *testing.T) {
	t.Run("yes", func(t *testing.T) {
		u := mustUser(t)
		assert.NoError(t, u.Verify())
		assert.True(t, u.CanLogin())
	})

	t.Run("no unverified", func(t *testing.T) {
		u := mustUser(t)
		assert.False(t, u.CanLogin())
	})

	t.Run("no banned", func(t *testing.T) {
		u := mustUser(t)
		_ = u.Verify()
		_ = u.Ban()
		assert.False(t, u.CanLogin())
	})

	t.Run("no uninitialized", func(t *testing.T) {
		var u domain.User
		assert.False(t, u.CanLogin())
	})
}

func TestUserVerify(t *testing.T) {
	t.Run("activated", func(t *testing.T) {
		u := mustUser(t)
		assert.NoError(t, u.Verify())
		assert.NotNil(t, u.VerifiedAt)
	})

	t.Run("already verified", func(t *testing.T) {
		u := mustUser(t)
		now := time.Now().UTC()
		u.VerifiedAt = &now
		assert.ErrorIs(t, u.Verify(), domain.ErrUserVerified)
	})

	t.Run("banned", func(t *testing.T) {
		u := mustUser(t)
		u.Status = domain.StatusBanned
		assert.ErrorIs(t, u.Verify(), domain.ErrUserNotActivated)
	})

	t.Run("uninitialized", func(t *testing.T) {
		var u domain.User
		assert.ErrorIs(t, u.Verify(), domain.ErrUserNotActivated)
	})
}

func TestUserBan(t *testing.T) {
	t.Run("activated", func(t *testing.T) {
		u := mustUser(t)
		assert.NoError(t, u.Ban())
		assert.True(t, u.IsBanned())
	})

	t.Run("already banned", func(t *testing.T) {
		u := mustUser(t)
		u.Status = domain.StatusBanned
		assert.ErrorIs(t, u.Ban(), domain.ErrUserBanned)
	})

	t.Run("deleted", func(t *testing.T) {
		u := mustUser(t)
		u.Status = domain.StatusDeleted
		assert.ErrorIs(t, u.Ban(), domain.ErrUserDeleted)
	})

	t.Run("uninitialized", func(t *testing.T) {
		var u domain.User
		assert.NoError(t, u.Ban())
		assert.True(t, u.IsBanned())
	})
}

func TestUserUnban(t *testing.T) {
	t.Run("banned", func(t *testing.T) {
		u := mustUser(t)
		u.Status = domain.StatusBanned
		assert.NoError(t, u.Unban())
		assert.True(t, u.IsActivated())
	})

	t.Run("activated", func(t *testing.T) {
		u := mustUser(t)
		assert.ErrorIs(t, u.Unban(), domain.ErrUserNotBanned)
	})

	t.Run("uninitialized", func(t *testing.T) {
		var u domain.User
		assert.ErrorIs(t, u.Unban(), domain.ErrUserNotBanned)
	})
}

func TestUserDelete(t *testing.T) {
	t.Run("activated", func(t *testing.T) {
		u := mustUser(t)
		assert.NoError(t, u.Delete())
		assert.True(t, u.IsDeleted())
	})

	t.Run("already deleted", func(t *testing.T) {
		u := mustUser(t)
		u.Status = domain.StatusDeleted
		assert.ErrorIs(t, u.Delete(), domain.ErrUserDeleted)
	})

	t.Run("uninitialized", func(t *testing.T) {
		var u domain.User
		assert.NoError(t, u.Delete())
		assert.True(t, u.IsDeleted())
	})
}

func TestUserChangePassword(t *testing.T) {
	newPass, _ := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=2$n$h")

	t.Run("ok", func(t *testing.T) {
		u := mustUser(t)
		assert.NoError(t, u.ChangePassword(newPass))
		v, _ := u.Password.Value()
		nv, _ := newPass.Value()
		assert.Equal(t, nv, v)
	})

	t.Run("banned", func(t *testing.T) {
		u := mustUser(t)
		u.Status = domain.StatusBanned
		assert.ErrorIs(t, u.ChangePassword(newPass), domain.ErrUserNotActivated)
	})

	t.Run("uninitialized", func(t *testing.T) {
		var u domain.User
		assert.ErrorIs(t, u.ChangePassword(newPass), domain.ErrUserNotActivated)
	})
}

func TestUserUpdateInfo(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		u := mustUser(t)
		assert.NoError(t, u.UpdateInfo("Jane", "Smith"))
		assert.Equal(t, "Jane", u.FirstName)
		assert.Equal(t, "Smith", u.LastName)
	})

	t.Run("empty first", func(t *testing.T) {
		u := mustUser(t)
		assert.ErrorIs(t, u.UpdateInfo("", userTestLastName), domain.ErrFirstNameRequired)
	})

	t.Run("empty last", func(t *testing.T) {
		u := mustUser(t)
		assert.ErrorIs(t, u.UpdateInfo(userTestFirstName, ""), domain.ErrLastNameRequired)
	})

	t.Run("banned", func(t *testing.T) {
		u := mustUser(t)
		u.Status = domain.StatusBanned
		assert.ErrorIs(t, u.UpdateInfo("Jane", "Smith"), domain.ErrUserNotActivated)
	})

	t.Run("uninitialized", func(t *testing.T) {
		var u domain.User
		assert.ErrorIs(t, u.UpdateInfo("Jane", "Smith"), domain.ErrUserNotActivated)
	})
}

func TestUserUpdateRole(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		u := mustUser(t)
		admin, _ := domain.NewRole(domain.RoleAdmin)
		assert.NoError(t, u.UpdateRole(admin))
		assert.Equal(t, admin, u.Role)
	})

	t.Run("banned", func(t *testing.T) {
		u := mustUser(t)
		u.Status = domain.StatusBanned
		admin, _ := domain.NewRole(domain.RoleAdmin)
		assert.ErrorIs(t, u.UpdateRole(admin), domain.ErrUserNotActivated)
	})

	t.Run("uninitialized", func(t *testing.T) {
		var u domain.User

		admin, _ := domain.NewRole(domain.RoleAdmin)
		assert.ErrorIs(t, u.UpdateRole(admin), domain.ErrUserNotActivated)
	})
}
