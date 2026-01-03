package logger

import (
	"fmt"
	"sync"
)

type Factory func(Config) (Logger, error)

var (
	mu       sync.RWMutex
	registry = make(map[Driver]Factory)
)

func Register(d Driver, f Factory) {
	mu.Lock()
	defer mu.Unlock()
	if f == nil {
		panic("logger: factory is nil")
	}
	registry[d] = f
}

func New(opts ...Option) (Logger, error) {
	cfg := Config{
		Driver:            DriverZap,
		Level:             LevelInfo,
		Encoding:          EncodingConsole,
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		OutputPaths:       []string{"stdout"},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	mu.RLock()
	factory, exists := registry[cfg.Driver]
	mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("logger: driver %q not found (forgot to import?)", cfg.Driver)
	}

	return factory(cfg)
}
