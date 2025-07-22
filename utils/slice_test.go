package utils_test

import (
	"testing"
	"time"

	"github.com/Piszmog/pathwise/types"
	"github.com/Piszmog/pathwise/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetFirstElementID(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name     string
		vals     []types.JobApplication
		expected int64
	}{
		{
			name:     "empty slice",
			vals:     []types.JobApplication{},
			expected: 0,
		},
		{
			name: "single element",
			vals: []types.JobApplication{
				{ID: 123, Company: "Test Company", CreatedAt: now},
			},
			expected: 123,
		},
		{
			name: "multiple elements",
			vals: []types.JobApplication{
				{ID: 456, Company: "First Company", CreatedAt: now},
				{ID: 789, Company: "Second Company", CreatedAt: now},
				{ID: 101, Company: "Third Company", CreatedAt: now},
			},
			expected: 456,
		},
		{
			name: "zero id first element",
			vals: []types.JobApplication{
				{ID: 0, Company: "Zero Company", CreatedAt: now},
				{ID: 123, Company: "Non-zero Company", CreatedAt: now},
			},
			expected: 0,
		},
		{
			name: "negative id first element",
			vals: []types.JobApplication{
				{ID: -123, Company: "Negative Company", CreatedAt: now},
				{ID: 456, Company: "Positive Company", CreatedAt: now},
			},
			expected: -123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.GetFirstElementID(tt.vals)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFirstElementType(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name     string
		vals     []types.JobApplicationTimelineEntry
		expected types.JobApplicationTimelineType
	}{
		{
			name:     "empty slice",
			vals:     []types.JobApplicationTimelineEntry{},
			expected: "",
		},
		{
			name: "single note element",
			vals: []types.JobApplicationTimelineEntry{
				types.JobApplicationNote{
					ID:        123,
					Note:      "Test note",
					CreatedAt: now,
				},
			},
			expected: types.JobApplicationTimelineTypeNote,
		},
		{
			name: "single status element",
			vals: []types.JobApplicationTimelineEntry{
				types.JobApplicationStatusHistory{
					ID:        456,
					Status:    types.JobApplicationStatusApplied,
					CreatedAt: now,
				},
			},
			expected: types.JobApplicationTimelineTypeStatus,
		},
		{
			name: "multiple elements - note first",
			vals: []types.JobApplicationTimelineEntry{
				types.JobApplicationNote{
					ID:        123,
					Note:      "First note",
					CreatedAt: now,
				},
				types.JobApplicationStatusHistory{
					ID:        456,
					Status:    types.JobApplicationStatusApplied,
					CreatedAt: now,
				},
			},
			expected: types.JobApplicationTimelineTypeNote,
		},
		{
			name: "multiple elements - status first",
			vals: []types.JobApplicationTimelineEntry{
				types.JobApplicationStatusHistory{
					ID:        789,
					Status:    types.JobApplicationStatusInterviewing,
					CreatedAt: now,
				},
				types.JobApplicationNote{
					ID:        101,
					Note:      "Second note",
					CreatedAt: now,
				},
			},
			expected: types.JobApplicationTimelineTypeStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.GetFirstElementType(tt.vals)
			assert.Equal(t, tt.expected, result)
		})
	}
}
