package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/internal/config"
)

const configYAML = `
app:
  name: go-auth-test
  env: dev
server:
  host: localhost
  port: 8080
  idle_to: 60s
  read_to: 10s
  write_to: 10s
  shutdown_to: 15s
cors:
  origins: [http://localhost:3000]
  methods: [GET, POST]
  headers: [Content-Type]
  max_age: 600
rate_limit:
  limit: 100
  period: 1m
database:
  name: testdb
  host: localhost
  port: 5432
  user: postgres
  password: password
  ssl_mode: disable
  max_conns: 10
  max_idle: 5
auth:
  jwt_secret: 12345678901234567890123456789012
  access_ttl: 15m
  refresh_ttl: 24h
  hash_cost: 10
smtp:
  host: smtp.mailtrap.io
  port: 2525
  username: "user"
  password: password
  from: no-reply@example.com
`

func TestLoadFromReader(t *testing.T) {
	type modifierFunc func(string) string

	tests := []struct {
		name     string
		modifier modifierFunc
		want     error
		assert   func(*testing.T, *config.Config)
	}{
		{
			name: "valid config",
			assert: func(t *testing.T, c *config.Config) {
				assert.Equal(t, "go-auth-test", c.App.Name)
				assert.Equal(t, 8080, int(c.Server.Port))
				assert.Equal(t, 60*time.Second, c.Server.IdleTO)
			},
		},
		{
			name: "required field missing",
			modifier: func(s string) string {
				return strings.Replace(s, `name: go-auth-test`, "", 1)
			},
			want: config.ErrConfigValidation,
		},
		{
			name: "readto greater than idleto",
			modifier: func(s string) string {
				s = strings.Replace(s, `idle_to: 60s`, `idle_to: 10s`, 1)
				s = strings.Replace(s, `read_to: 10s`, `read_to: 20s`, 1)

				return s
			},
			want: config.ErrConfigValidation,
		},
		{
			name: "syntax invalid yaml",
			modifier: func(s string) string {
				return strings.Replace(s, `port: 8080`, `port: "not-a-number"`, 1)
			},
			want: config.ErrConfigUnmarshal,
		},
		{
			name: "invalid enum",
			modifier: func(s string) string {
				return strings.Replace(s, `env: dev`, `env: staging`, 1)
			},
			want: config.ErrConfigValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			yamlContent := configYAML
			if tt.modifier != nil {
				yamlContent = tt.modifier(yamlContent)
			}

			l := config.NewLoader()
			cfg, err := l.LoadFromReader(strings.NewReader(yamlContent))

			if tt.want != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.want)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)

				if tt.assert != nil {
					tt.assert(t, cfg)
				}
			}
		})
	}
}

func TestLoadFileNotFound(t *testing.T) {
	l := config.NewLoader()

	cfg, err := l.Load("")

	require.Error(t, err)
	assert.Nil(t, cfg)

	assert.ErrorIs(t, err, config.ErrConfigNotFound)
}
