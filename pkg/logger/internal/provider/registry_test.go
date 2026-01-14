package provider_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/provider"
)

// mockFactory is a simple factory for testing purposes.
func mockFactory(_ *core.Config) (provider.Logger, error) {
	return nil, nil
}

func TestRegistry(t *testing.T) {
	t.Run("success register and get", func(t *testing.T) {
		driverName := core.Driver("test_driver_success")
		provider.Register(driverName, mockFactory)
		factory, err := provider.Get(driverName)

		require.NoError(t, err)
		assert.NotNil(t, factory)
	})

	t.Run("nil factory panic", func(t *testing.T) {
		driverName := core.Driver("test_driver_nil")

		assert.PanicsWithError(t, core.ErrNilFactory.Error(), func() {
			provider.Register(driverName, nil)
		})
	})

	t.Run("duplicate registration panic", func(t *testing.T) {
		driverName := core.Driver("test_driver_duplicate")
		provider.Register(driverName, mockFactory)

		assert.PanicsWithError(t, core.ErrFactoryAlreadyRegistered.Error(), func() {
			provider.Register(driverName, mockFactory)
		})
	})

	t.Run("unknown driver", func(t *testing.T) {
		driverName := core.Driver("unknown_driver")
		factory, err := provider.Get(driverName)
		assert.ErrorIs(t, err, core.ErrUnknownDriver)
		assert.Nil(t, factory)
	})
}

func TestRegistryConcurrency(t *testing.T) {
	// Test concurrent access to Register and Get
	var wg sync.WaitGroup

	workers := 20

	// Start workers that register unique drivers
	for i := range workers {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			driverName := core.Driver("concurrent_driver_" + string(rune(id)))
			provider.Register(driverName, mockFactory)
		}(i)
	}

	// Start workers that read from registry
	for range workers {
		wg.Go(func() {
			// Try to get a potentially existing driver
			_, _ = provider.Get(core.Driver("concurrent_driver_0"))
			// Try to get a non-existing driver
			_, _ = provider.Get(core.Driver("non_existent"))
		})
	}

	wg.Wait()
}
