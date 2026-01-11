package zap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToZapFields(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected []any
		dirty    bool
	}{
		{
			name: "nil input",
		},
		{
			name:     "empty input",
			input:    []any{},
			expected: []any{},
		},
		{
			name:     "happy path",
			input:    []any{"key1", "val1", "key2", 123},
			expected: []any{"key1", "val1", "key2", 123},
		},
		{
			name:     "odd number of arguments",
			input:    []any{"key1", "val1", "key2"},
			expected: []any{"key1", "val1", "key2", "_MISSING_"},
			dirty:    true,
		},
		{
			name:     "non-string key",
			input:    []any{"key1", "val1", 123, "val2"},
			expected: []any{"key1", "val1", "123", "val2"},
			dirty:    true,
		},
		{
			name:     "reserved key caller",
			input:    []any{"caller", "main.go:123", "key1", "val1"},
			expected: []any{"field.caller", "main.go:123", "key1", "val1"},
			dirty:    true,
		},
		{
			name:     "reserved key msg",
			input:    []any{"msg", "hello"},
			expected: []any{"field.msg", "hello"},
			dirty:    true,
		},
		{
			name:     "mixed dirty case",
			input:    []any{123, "val1", "level", "val2", "key3"},
			expected: []any{"123", "val1", "field.level", "val2", "key3", "_MISSING_"},
			dirty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toZapFields(tt.input)
			assert.Equal(t, tt.expected, result)

			if !tt.dirty && len(tt.input) > 0 {
				// Optimization check: pointers should be same if not dirty
				assert.Same(t, &tt.input[0], &result[0], "slice reallocation should not happen for clean input")
			}
		})
	}
}

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{input: "simple", expected: "simple"},
		{input: 123, expected: "123"},
		{input: true, expected: "true"},
		{input: "level", expected: "field.level"},
		{input: "msg", expected: "field.msg"},
		{input: "time", expected: "field.time"},
		{input: "caller", expected: "field.caller"},
		{input: "stacktrace", expected: "field.stacktrace"},
		{input: "other", expected: "other"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeKey(tt.input))
		})
	}
}
