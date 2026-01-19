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
	envPrefix = "GO_AUTH"
)

var (
	ErrConfigNotFound   = errors.New("config: file not found")
	ErrConfigRead       = errors.New("config: read error")
	ErrConfigUnmarshal  = errors.New("config: unmarshal error")
	ErrConfigValidation = errors.New("config: validation error")
)

func NewLoader() *Loader {
	vip := viper.New()

	vip.AddConfigPath(filePath)
	vip.SetConfigName(fileName)
	vip.SetConfigType(fileType)

	vip.SetEnvPrefix(envPrefix)
	vip.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vip.AutomaticEnv()

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

	return l.process()
}

func (l *Loader) LoadFromReader(r io.Reader) (*Config, error) {
	if err := l.vip.ReadConfig(r); err != nil {
		return nil, wrapError(ErrConfigRead, err)
	}

	return l.process()
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
