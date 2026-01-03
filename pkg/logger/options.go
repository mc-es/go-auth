package logger

type Option func(*Config)

func WithDriver(d Driver) Option {
	return func(c *Config) {
		c.Driver = d
	}
}

func WithLevel(l Level) Option {
	return func(c *Config) {
		c.Level = l
	}
}

func WithEncoding(e Encoding) Option {
	return func(c *Config) {
		c.Encoding = e
	}
}

func WithDevelopment() Option {
	return func(c *Config) {
		c.Development = true
	}
}

func WithDisableCaller() Option {
	return func(c *Config) {
		c.DisableCaller = true
	}
}

func WithDisableStacktrace() Option {
	return func(c *Config) {
		c.DisableStacktrace = true
	}
}

func WithOutputPaths(paths ...string) Option {
	return func(c *Config) {
		c.OutputPaths = paths
	}
}

func WithInitialFields(fields map[string]any) Option {
	return func(c *Config) {
		c.InitialFields = fields
	}
}

func WithValue(key string, value any) Option {
	return func(c *Config) {
		if c.DriverOptions == nil {
			c.DriverOptions = make(map[string]any)
		}
		c.DriverOptions[key] = value
	}
}
