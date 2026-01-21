package registry

import (
	"sync"

	"go-auth/pkg/logger/internal/core"
)

type Factory func(config *core.Config) (core.Logger, error)

var (
	mu        sync.RWMutex
	factories = make(map[core.Driver]Factory)
)

func Register(driver core.Driver, factory Factory) {
	mu.Lock()
	defer mu.Unlock()

	if factory == nil {
		panic(core.ErrNilFactory)
	}

	if _, ok := factories[driver]; ok {
		panic(core.ErrFactoryAlreadyRegistered)
	}

	factories[driver] = factory
}

func Get(driver core.Driver) (Factory, error) {
	mu.RLock()
	defer mu.RUnlock()

	factory, ok := factories[driver]
	if !ok {
		return nil, core.ErrUnknownDriver
	}

	return factory, nil
}
