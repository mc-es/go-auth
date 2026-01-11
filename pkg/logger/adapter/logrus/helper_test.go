package logrus

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestToLogrusFields(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected logrus.Fields
	}{
		{
			name:     "nil input",
			expected: logrus.Fields{},
		},
		{
			name:     "empty input",
			input:    []any{},
			expected: logrus.Fields{},
		},
		{
			name:     "happy path",
			input:    []any{"key1", "val1", "key2", 123},
			expected: logrus.Fields{"key1": "val1", "key2": 123},
		},
		{
			name:     "odd number of arguments",
			input:    []any{"key1", "val1", "key2"},
			expected: logrus.Fields{"key1": "val1", "key2": "_MISSING_"},
		},
		{
			name:     "non-string key",
			input:    []any{"key1", "val1", 123, "val2"},
			expected: logrus.Fields{"key1": "val1", "123": "val2"},
		},
		{
			name:     "reserved key caller",
			input:    []any{"caller", "main.go:123", "key1", "val1"},
			expected: logrus.Fields{"field.caller": "main.go:123", "key1": "val1"},
		},
		{
			name:     "reserved key msg",
			input:    []any{"msg", "hello"},
			expected: logrus.Fields{"field.msg": "hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, toLogrusFields(tt.input))
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
		{input: "logrus_error", expected: "field.logrus_error"},
		{input: "other", expected: "other"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeKey(tt.input))
		})
	}
}
