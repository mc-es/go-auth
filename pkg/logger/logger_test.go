package logger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go-auth/pkg/logger"
	"go-auth/pkg/logger/internal/core"
)

func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		opts        []logger.Option
		expectedErr error
	}{
		{
			name:        "missing driver",
			opts:        []logger.Option{logger.WithDriver(logger.Driver(""))},
			expectedErr: core.ErrMissingDriver,
		},
		{
			name:        "unknown driver",
			opts:        []logger.Option{logger.WithDriver(logger.Driver("unknown"))},
			expectedErr: core.ErrUnknownDriver,
		},
		{
			name:        "invalid level",
			opts:        []logger.Option{logger.WithLevel(logger.Level(99))},
			expectedErr: core.ErrInvalidLevel,
		},
		{
			name:        "invalid format",
			opts:        []logger.Option{logger.WithFormat(logger.Format("xml"))},
			expectedErr: core.ErrInvalidFormat,
		},
		{
			name:        "invalid time layout",
			opts:        []logger.Option{logger.WithTimeLayout(logger.TimeLayout(""))},
			expectedErr: core.ErrInvalidTimeLayout,
		},
		{
			name:        "empty output paths",
			opts:        []logger.Option{logger.WithOutputPaths()},
			expectedErr: core.ErrInvalidPaths,
		},
		{
			name:        "invalid output paths",
			opts:        []logger.Option{logger.WithOutputPaths("")},
			expectedErr: core.ErrInvalidPaths,
		},
		{
			name:        "invalid file rotation",
			opts:        []logger.Option{logger.WithFileRotation(0, -1, -100, false, false)},
			expectedErr: core.ErrInvalidFileRotation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := logger.New(tt.opts...)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Nil(t, log)
		})
	}
}
