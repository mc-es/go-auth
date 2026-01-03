package logger

// Option represents a logger configuration option.
type Option func(*Config)

// WithDriver sets the driver of the logger.
func WithDriver(driver driver) Option {
	return func(c *Config) {
		c.Driver = driver
	}
}

// WithLevel sets the level of the logger.
func WithLevel(level level) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithFormatter sets the formatter of the logger.
func WithFormatter(formatter formatter) Option {
	return func(c *Config) {
		c.Formatter = formatter
	}
}

// WithTimeLayout sets the time layout of the logger.
func WithTimeLayout(layout timeLayout) Option {
	return func(c *Config) {
		c.TimeLayout = layout
	}
}

// WithOutputPath sets the output path of the logger.
func WithOutputPath(paths ...string) Option {
	return func(c *Config) {
		c.OutputPath = paths
	}
}

// WithDevMode sets the development mode of the logger.
func WithDevMode() Option {
	return func(c *Config) {
		c.DevMode = true
	}
}

// WithoutCaller sets the caller mode of the logger.
func WithoutCaller() Option {
	return func(c *Config) {
		c.DisableCaller = true
	}
}
