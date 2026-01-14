package output_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/output"
)

const (
	testLogFileName = "test.log"
	logsDir         = "logs"
)

// setupDestination creates a destination with the given config and returns it.
func setupDestination(t *testing.T, cfg *core.Config) *output.Destination {
	t.Helper()

	dest, err := output.New(cfg)
	require.NoError(t, err)
	require.NotNil(t, dest)

	return dest
}

// createTempFile returns a temporary file path.
func createTempFile(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), testLogFileName)
}

func TestNewStdout(t *testing.T) {
	cfg := &core.Config{
		OutputPaths: []string{"stdout"},
	}

	dest := setupDestination(t, cfg)

	defer func() { _ = dest.Close() }()

	assert.NotNil(t, dest.Console)
	assert.Nil(t, dest.File)
}

func TestNewStderr(t *testing.T) {
	cfg := &core.Config{
		OutputPaths: []string{"stderr"},
	}

	dest := setupDestination(t, cfg)

	defer func() { _ = dest.Close() }()

	assert.NotNil(t, dest.Console)
	assert.Nil(t, dest.File)
}

func TestNewFile(t *testing.T) {
	logFile := createTempFile(t)

	cfg := &core.Config{
		OutputPaths: []string{logFile},
	}

	dest := setupDestination(t, cfg)

	defer func() { _ = dest.Close() }()

	assert.Nil(t, dest.Console)
	assert.NotNil(t, dest.File)
}

func TestNewStdoutAndFile(t *testing.T) {
	logFile := createTempFile(t)

	cfg := &core.Config{
		OutputPaths: []string{"stdout", logFile},
	}

	dest := setupDestination(t, cfg)

	defer func() { _ = dest.Close() }()

	assert.NotNil(t, dest.Console)
	assert.NotNil(t, dest.File)
}

func TestNewDuplicatePaths(t *testing.T) {
	logFile := createTempFile(t)

	cfg := &core.Config{
		OutputPaths: []string{"stdout", "stdout", logFile, logFile},
	}

	dest := setupDestination(t, cfg)

	defer func() { _ = dest.Close() }()

	assert.NotNil(t, dest.Console)
	assert.NotNil(t, dest.File)
}

func TestNewFileRotation(t *testing.T) {
	logFile := createTempFile(t)

	cfg := &core.Config{
		OutputPaths: []string{logFile},
		FileRotation: core.FileRotation{
			MaxAge:     7,
			MaxSize:    100,
			MaxBackups: 3,
			LocalTime:  true,
			Compress:   true,
		},
	}

	dest := setupDestination(t, cfg)

	defer func() { _ = dest.Close() }()

	assert.NotNil(t, dest.File)
}

func TestNewDirectoryAsPath(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &core.Config{
		OutputPaths: []string{tempDir}, // Directory, not file
	}

	dest, err := output.New(cfg)
	assert.Error(t, err)
	assert.Nil(t, dest)
	assert.Contains(t, err.Error(), "directory")
}

func TestNewCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, logsDir)
	logFile := filepath.Join(logDir, testLogFileName)

	cfg := &core.Config{
		OutputPaths: []string{logFile},
	}

	dest := setupDestination(t, cfg)

	defer func() { _ = dest.Close() }()

	// Verify directory was created
	_, err := os.Stat(logDir)
	assert.NoError(t, err)
}
