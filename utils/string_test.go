package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanJobTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "leading quote only",
			input:    "\"Job title",
			expected: "Job title",
		},
		{
			name:     "wrapped in quotes",
			input:    "\"Job title\"",
			expected: "Job title",
		},
		{
			name:     "leading space",
			input:    " Job title",
			expected: "Job title",
		},
		{
			name:     "leading quote and space",
			input:    "\" Job title",
			expected: "Job title",
		},
		{
			name:     "trailing space",
			input:    "Job title ",
			expected: "Job title",
		},
		{
			name:     "leading and trailing spaces",
			input:    " Job title ",
			expected: "Job title",
		},
		{
			name:     "normal title",
			input:    "Job title",
			expected: "Job title",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only quotes",
			input:    "\"\"",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
		{
			name:     "complex case with multiple issues",
			input:    " \"Senior Software Engineer\" ",
			expected: "Senior Software Engineer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanJobTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
