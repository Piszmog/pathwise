package utils_test

import (
	"testing"

	"github.com/Piszmog/pathwise/utils"
	"github.com/stretchr/testify/assert"
)

func TestJobRowID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		id       int64
		expected string
	}{
		{
			name:     "positive id",
			id:       123,
			expected: "job-123-row",
		},
		{
			name:     "zero id",
			id:       0,
			expected: "job-0-row",
		},
		{
			name:     "negative id",
			id:       -456,
			expected: "job--456-row",
		},
		{
			name:     "large id",
			id:       9223372036854775807, // max int64
			expected: "job-9223372036854775807-row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.JobRowID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJobRowMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		id       int64
		expected string
	}{
		{
			name:     "positive id",
			id:       123,
			expected: "job-123-row-metadata",
		},
		{
			name:     "zero id",
			id:       0,
			expected: "job-0-row-metadata",
		},
		{
			name:     "negative id",
			id:       -456,
			expected: "job--456-row-metadata",
		},
		{
			name:     "large id",
			id:       9223372036854775807, // max int64
			expected: "job-9223372036854775807-row-metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.JobRowMetadata(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimelineStatusRowID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		id       int64
		expected string
	}{
		{
			name:     "positive id",
			id:       789,
			expected: "timeline-status-789-row",
		},
		{
			name:     "zero id",
			id:       0,
			expected: "timeline-status-0-row",
		},
		{
			name:     "negative id",
			id:       -123,
			expected: "timeline-status--123-row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.TimelineStatusRowID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimelineStatusRowStringID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "numeric string",
			id:       "123",
			expected: "timeline-status-123-row",
		},
		{
			name:     "empty string",
			id:       "",
			expected: "timeline-status--row",
		},
		{
			name:     "alphanumeric string",
			id:       "abc123",
			expected: "timeline-status-abc123-row",
		},
		{
			name:     "string with special characters",
			id:       "test-id_123",
			expected: "timeline-status-test-id_123-row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.TimelineStatusRowStringID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimelineNoteRowID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		id       int64
		expected string
	}{
		{
			name:     "positive id",
			id:       456,
			expected: "timeline-note-456-row",
		},
		{
			name:     "zero id",
			id:       0,
			expected: "timeline-note-0-row",
		},
		{
			name:     "negative id",
			id:       -789,
			expected: "timeline-note--789-row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.TimelineNoteRowID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimelineNoteRowStringID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "numeric string",
			id:       "456",
			expected: "timeline-note-456-row",
		},
		{
			name:     "empty string",
			id:       "",
			expected: "timeline-note--row",
		},
		{
			name:     "alphanumeric string",
			id:       "note123",
			expected: "timeline-note-note123-row",
		},
		{
			name:     "string with special characters",
			id:       "note-id_456",
			expected: "timeline-note-note-id_456-row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.TimelineNoteRowStringID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}
