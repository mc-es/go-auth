package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Loader struct {
	v          *viper.Viper
	configName string
	configPath string
	configType string
	envPrefix  string
}

type Option func(*Loader)

func NewLoader(opts ...Option) *Loader {
	loader := &Loader{
		v:          viper.New(),
		configName: "config",
		configPath: ".",
		configType: "yaml",
		envPrefix:  "",
	}

	for _, opt := range opts {
		opt(loader)
	}

	loader.v.SetEnvPrefix(loader.envPrefix)
	loader.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	loader.v.AutomaticEnv()

	return loader
}

func (l *Loader) Load(envFile string) (*Config, error) {
	_ = godotenv.Load(envFile)

	l.v.SetConfigName(l.configName)
	l.v.AddConfigPath(l.configPath)
	l.v.SetConfigType(l.configType)

	if err := l.v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("config file read error: %w", err)
		}
	}

	var cfg Config
	if err := l.v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config parse error: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return &cfg, nil
}

func WithConfigName(name string) Option {
	return func(l *Loader) {
		l.configName = name
	}
}

func WithConfigPath(path string) Option {
	return func(l *Loader) {
		l.configPath = path
	}
}

func WithConfigType(configType string) Option {
	return func(l *Loader) {
		l.configType = configType
	}
}

func WithEnvPrefix(prefix string) Option {
	return func(l *Loader) {
		l.envPrefix = prefix
	}
}
