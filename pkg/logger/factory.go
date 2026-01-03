package logger

import (
	"fmt"
	"sync"
)

type factory func(config *Config) Logger

var (
	mu       sync.RWMutex
	adapters = make(map[driver]factory)
)

// Register registers a new logger driver.
func Register(driver driver, factory factory) {
	mu.Lock()
	defer mu.Unlock()

	if factory == nil {
		panic("logger: factory is nil")
	}

	adapters[driver] = factory
}

// New creates a new logger instance.
func New(opts ...Option) (Logger, error) {
	cfg := Config{
		Driver:        DriverZap,
		Level:         LevelDebug,
		Formatter:     FormatterJSON,
		TimeLayout:    TimeLayoutDateTime,
		OutputPath:    []string{"stdout"},
		DevMode:       false,
		DisableCaller: false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	mu.RLock()

	factory, exists := adapters[cfg.Driver]

	mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("logger: driver %q not found (forgot to import?)", cfg.Driver)
	}

	return factory(&cfg), nil
}
