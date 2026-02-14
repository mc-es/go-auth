package config

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Loader struct {
	vip *viper.Viper
	val *validator.Validate
}

const (
	filePath  = "."
	fileName  = ".app-config"
	fileType  = "yaml"
	envPrefix = ""
)

var (
	ErrConfigNotFound   = errors.New("config: file not found")
	ErrConfigRead       = errors.New("config: read error")
	ErrConfigUnmarshal  = errors.New("config: unmarshal error")
	ErrConfigValidation = errors.New("config: validation error")
	ErrSensitiveConfig  = errors.New("config: sensitive keys found in yaml file, use env vars instead")
)

var forbiddenFileKeys = []string{
	"database.name",
	"database.host",
	"database.port",
	"database.user",
	"database.password",
	"database.sslmode",
	"security.jwt_secret",
	"smtp.host",
	"smtp.port",
	"smtp.username",
	"smtp.password",
	"smtp.from",
}

func NewLoader() *Loader {
	vip := viper.New()

	vip.AddConfigPath(filePath)
	vip.SetConfigName(fileName)
	vip.SetConfigType(fileType)

	vip.SetEnvPrefix(envPrefix)
	vip.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vip.AutomaticEnv()

	for _, key := range forbiddenFileKeys {
		_ = vip.BindEnv(key)
	}

	return &Loader{
		vip: vip,
		val: validator.New(),
	}
}

func (l *Loader) Load(envFile string) (*Config, error) {
	if envFile != "" {
		_ = godotenv.Load(envFile)
	}

	if err := l.vip.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return nil, wrapError(ErrConfigNotFound, err)
		}

		return nil, wrapError(ErrConfigRead, err)
	}

	if err := l.ensureNoSensitiveKeysInFile(); err != nil {
		return nil, err
	}

	return l.process()
}

func (l *Loader) LoadFromReader(r io.Reader) (*Config, error) {
	if err := l.vip.ReadConfig(r); err != nil {
		return nil, wrapError(ErrConfigRead, err)
	}

	if err := l.ensureNoSensitiveKeysInFile(); err != nil {
		return nil, err
	}

	return l.process()
}

func (l *Loader) ensureNoSensitiveKeysInFile() error {
	for _, key := range forbiddenFileKeys {
		if l.vip.InConfig(key) {
			return fmt.Errorf("%w: key '%s' is not allowed in %s.%s", ErrSensitiveConfig, key, fileName, fileType)
		}
	}

	return nil
}

func (l *Loader) process() (*Config, error) {
	var cfg Config
	if err := l.vip.Unmarshal(&cfg); err != nil {
		return nil, wrapError(ErrConfigUnmarshal, err)
	}

	if err := l.val.Struct(&cfg); err != nil {
		return nil, wrapError(ErrConfigValidation, err)
	}

	return &cfg, nil
}

func wrapError(sentinel, err error) error {
	return fmt.Errorf("%w: %v", sentinel, err)
}
