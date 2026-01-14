package core

import "errors"

var (
	ErrMissingDriver       = errors.New("logger: driver is missing")
	ErrUnknownDriver       = errors.New("logger: unknown driver (forgot to import?)")
	ErrInvalidLevel        = errors.New("logger: invalid log level")
	ErrInvalidFormat       = errors.New("logger: invalid log format")
	ErrInvalidTimeLayout   = errors.New("logger: invalid time layout")
	ErrInvalidPaths        = errors.New("logger: invalid output paths")
	ErrInvalidFileRotation = errors.New("logger: invalid file rotation")

	ErrNilFactory               = errors.New("logger: factory is nil")
	ErrFactoryAlreadyRegistered = errors.New("logger: factory already registered")
)
