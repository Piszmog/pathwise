package utils_test

import (
	"testing"

	"github.com/Piszmog/pathwise/ui/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetActualMin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		start    int64
		total    int64
		expected int64
	}{
		{
			name:     "start less than total",
			start:    5,
			total:    10,
			expected: 5,
		},
		{
			name:     "start equal to total",
			start:    10,
			total:    10,
			expected: 10,
		},
		{
			name:     "start greater than total",
			start:    15,
			total:    10,
			expected: 10,
		},
		{
			name:     "zero start",
			start:    0,
			total:    5,
			expected: 0,
		},
		{
			name:     "zero total",
			start:    5,
			total:    0,
			expected: 0,
		},
		{
			name:     "both zero",
			start:    0,
			total:    0,
			expected: 0,
		},
		{
			name:     "negative start",
			start:    -5,
			total:    10,
			expected: -5,
		},
		{
			name:     "negative total",
			start:    5,
			total:    -10,
			expected: -10,
		},
		{
			name:     "both negative",
			start:    -15,
			total:    -10,
			expected: -15,
		},
		{
			name:     "large numbers",
			start:    1000000,
			total:    2000000,
			expected: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.GetActualMin(tt.start, tt.total)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetActualMax(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		end      int64
		total    int64
		expected int64
	}{
		{
			name:     "end less than total",
			end:      5,
			total:    10,
			expected: 5,
		},
		{
			name:     "end equal to total",
			end:      10,
			total:    10,
			expected: 10,
		},
		{
			name:     "end greater than total",
			end:      15,
			total:    10,
			expected: 10,
		},
		{
			name:     "zero end",
			end:      0,
			total:    5,
			expected: 0,
		},
		{
			name:     "zero total",
			end:      5,
			total:    0,
			expected: 0,
		},
		{
			name:     "both zero",
			end:      0,
			total:    0,
			expected: 0,
		},
		{
			name:     "negative end",
			end:      -5,
			total:    10,
			expected: -5,
		},
		{
			name:     "negative total",
			end:      5,
			total:    -10,
			expected: -10,
		},
		{
			name:     "both negative",
			end:      -5,
			total:    -10,
			expected: -10,
		},
		{
			name:     "large numbers",
			end:      2000000,
			total:    1000000,
			expected: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.GetActualMax(tt.end, tt.total)
			assert.Equal(t, tt.expected, result)
		})
	}
}
