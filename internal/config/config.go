package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App       App       `mapstructure:"app"`
	Server    Server    `mapstructure:"server"`
	CORS      CORS      `mapstructure:"cors"`
	RateLimit RateLimit `mapstructure:"rate_limit"`
	Database  Database  `mapstructure:"database"`
	Auth      Auth      `mapstructure:"auth"`
	SMTP      SMTP      `mapstructure:"smtp"`
	Logger    Logger    `mapstructure:"logger"`
}

type App struct {
	Name string `mapstructure:"name" validate:"required,min=3,max=100"`
	Env  string `mapstructure:"env"  validate:"required,oneof=dev prod"`
}

type Server struct {
	Host       string        `mapstructure:"host"        validate:"required,hostname|ip"`
	Port       uint16        `mapstructure:"port"        validate:"required,port"`
	IdleTO     time.Duration `mapstructure:"idle_to"     validate:"required,min=5s,max=60s"`
	ReadTO     time.Duration `mapstructure:"read_to"     validate:"required,min=1s,max=30s,ltfield=IdleTO"`
	WriteTO    time.Duration `mapstructure:"write_to"    validate:"required,min=1s,max=30s,ltfield=IdleTO"`
	ShutdownTO time.Duration `mapstructure:"shutdown_to" validate:"required,min=1s,max=30s,gtefield=WriteTO"`
}

type CORS struct {
	Origins []string `mapstructure:"origins" validate:"required,min=1,dive,http_url|https_url"`
	Methods []string `mapstructure:"methods" validate:"required,min=1,dive,oneof=GET POST PUT DELETE OPTIONS"`
	Headers []string `mapstructure:"headers" validate:"required,min=1,dive,oneof=Content-Type Authorization"`
	MaxAge  int      `mapstructure:"max_age" validate:"required,min=0,max=86400"`
}

type RateLimit struct {
	Limit  int           `mapstructure:"limit"  validate:"required,min=1,max=1000"`
	Period time.Duration `mapstructure:"period" validate:"required,min=1s,max=1h"`
}

type Database struct {
	Name     string `mapstructure:"name"      validate:"required,min=3,max=100"`
	Host     string `mapstructure:"host"      validate:"required,hostname|ip"`
	Port     uint16 `mapstructure:"port"      validate:"required,port"`
	User     string `mapstructure:"user"      validate:"required"`
	Password string `mapstructure:"password"  validate:"required"`
	SSLMode  string `mapstructure:"ssl_mode"  validate:"required,oneof=disable require verify-ca verify-full"`
	MaxConns int    `mapstructure:"max_conns" validate:"required,min=1,max=100"`
	MaxIdle  int    `mapstructure:"max_idle"  validate:"required,min=1,max=50,ltefield=MaxConns"`
}

type Auth struct {
	JWTSecret  string        `mapstructure:"jwt_secret"  validate:"required,min=32,max=512"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"  validate:"required,min=5m,max=1h"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl" validate:"required,min=1h,max=168h,gtfield=AccessTTL"`
	HashCost   int           `mapstructure:"hash_cost"   validate:"required,min=10,max=15"`
}

type SMTP struct {
	Host     string `mapstructure:"host"     validate:"required,hostname|ip"`
	Port     uint16 `mapstructure:"port"     validate:"required,port"`
	Username string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	From     string `mapstructure:"from"     validate:"required,email"`
}

type Logger struct {
	Driver       string       `mapstructure:"driver"        validate:"oneof=logrus zap zerolog"`
	Level        string       `mapstructure:"level"         validate:"oneof=debug info warn error panic fatal"`
	Format       string       `mapstructure:"format"        validate:"oneof=json text"`
	TimeLayout   string       `mapstructure:"time_layout"   validate:"oneof=datetime date time rfc3339 rfc822 rfc1123"`
	OutputPaths  []string     `mapstructure:"output_paths"  validate:"min=1,dive"`
	Development  bool         `mapstructure:"development"`
	FileRotation FileRotation `mapstructure:"file_rotation"`
}

type FileRotation struct {
	MaxAge     int  `mapstructure:"max_age"     validate:"min=1,max=30"`
	MaxSize    int  `mapstructure:"max_size"    validate:"min=1,max=100"`
	MaxBackups int  `mapstructure:"max_backups" validate:"min=1,max=100"`
	LocalTime  bool `mapstructure:"local_time"`
	Compress   bool `mapstructure:"compress"`
}

func (c *Config) ServerAddr() string {
	return net.JoinHostPort(c.Server.Host, strconv.Itoa(int(c.Server.Port)))
}

func (c *Config) SMTPAddr() string {
	return net.JoinHostPort(c.SMTP.Host, strconv.Itoa(int(c.SMTP.Port)))
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.Name, c.Database.SSLMode,
	)
}

func (c *Config) DatabaseURL() string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.Database.User, c.Database.Password),
		Host:   net.JoinHostPort(c.Database.Host, strconv.Itoa(int(c.Database.Port))),
		Path:   c.Database.Name,
	}

	q := dsn.Query()
	q.Set("sslmode", c.Database.SSLMode)
	dsn.RawQuery = q.Encode()

	return dsn.String()
}

func (c *Config) JWTKey() []byte {
	return []byte(c.Auth.JWTSecret)
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Env, "prod")
}

func (c *Config) IsDevelopment() bool {
	return strings.EqualFold(c.App.Env, "dev")
}
