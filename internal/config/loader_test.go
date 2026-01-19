package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/internal/config"
)

const (
	testdataPath  = "testdata"
	validFile     = "config.valid"
	invalidFile   = "config.invalid"
	readErrorFile = "config.readerror"
	unmarshalFile = "config.unmarshal"
)

func newTestLoader(configName string, opts ...config.Option) *config.Loader {
	base := []config.Option{
		config.WithConfigName(configName),
		config.WithConfigPath(testdataPath),
		config.WithConfigType("yml"),
	}

	return config.NewLoader(append(base, opts...)...)
}

func TestLoadWithValidConfig(t *testing.T) {
	t.Parallel()

	l := newTestLoader(validFile)
	cfg, err := l.Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "go-auth", cfg.App.Name)
	assert.Equal(t, "dev", cfg.App.Env)
	assert.Equal(t, "0.0.0.0:8080", cfg.ServerAddr())
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "smtp.example.com:587", cfg.SMTPAddr())
	assert.Len(t, cfg.JWTKey(), 32)
}

func TestLoadWithEnvOverridesFile(t *testing.T) {
	t.Setenv("TEST_APP_NAME", "from-env")

	l := newTestLoader(validFile, config.WithEnvPrefix("TEST"))
	cfg, err := l.Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "from-env", cfg.App.Name)
}

func TestLoadWithInvalidConfig(t *testing.T) {
	t.Parallel()

	l := newTestLoader(invalidFile)
	cfg, err := l.Load("")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, config.ErrConfigValidation)
}

func TestLoadWithFileNotFound(t *testing.T) {
	t.Parallel()

	l := newTestLoader("nonexistent")
	cfg, err := l.Load("")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, config.ErrConfigNotFound)
}

func TestLoadWithReadError(t *testing.T) {
	t.Parallel()

	l := newTestLoader(readErrorFile)
	cfg, err := l.Load("")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, config.ErrConfigRead)
}

func TestLoadWithUnmarshalError(t *testing.T) {
	t.Parallel()

	l := newTestLoader(unmarshalFile)
	cfg, err := l.Load("")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, config.ErrConfigUnmarshal)
}
