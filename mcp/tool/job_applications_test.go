//go:build integration

package tool_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	contextkey "github.com/Piszmog/pathwise/internal/context_key"
	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/Piszmog/pathwise/mcp/tool"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/tursodatabase/go-libsql"
)

var errAuthenticationFailed = errors.New("authentication failed")

type testCase struct {
	name          string
	userID        int64
	setupData     func(t *testing.T, db *sql.DB, userID int64)
	expectedCount int
	expectedError error
}

func TestJobApplicationsTool(t *testing.T) {
	tests := []testCase{
		{
			name:          "no applications",
			userID:        1,
			setupData:     func(t *testing.T, db *sql.DB, userID int64) {},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name:   "single application",
			userID: 1,
			setupData: func(t *testing.T, db *sql.DB, userID int64) {
				insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
			},
			expectedCount: 1,
			expectedError: nil,
		},
		{
			name:   "multiple applications",
			userID: 1,
			setupData: func(t *testing.T, db *sql.DB, userID int64) {
				insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
				insertJobApplication(t, db, userID, "Company B", "Developer", "interviewing")
				insertJobApplication(t, db, userID, "Company C", "Manager", "rejected")
			},
			expectedCount: 3,
			expectedError: nil,
		},
		{
			name:   "user isolation - only sees own applications",
			userID: 1,
			setupData: func(t *testing.T, db *sql.DB, userID int64) {
				createTestUser(t, db, 2)
				insertJobApplication(t, db, 2, "Other Company", "Other Job", "applied")
				insertJobApplication(t, db, userID, "My Company", "My Job", "applied")
			},
			expectedCount: 1,
			expectedError: nil,
		},
		{
			name:          "unauthenticated user",
			userID:        0,
			setupData:     func(t *testing.T, db *sql.DB, userID int64) {},
			expectedCount: 0,
			expectedError: errAuthenticationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			database := setupTestDB(t)
			defer cleanupTestDB(t, database)

			handler := &tool.Handler{
				Logger:   setupTestLogger(),
				Database: database,
			}

			if tt.userID > 0 {
				createTestUser(t, database.DB(), tt.userID)
				tt.setupData(t, database.DB(), tt.userID)
			}

			jobAppsTool := handler.NewJobApplicationsTool()

			var ctx context.Context
			if tt.userID > 0 {
				ctx = context.WithValue(context.Background(), contextkey.KeyUserID, tt.userID)
			} else {
				ctx = context.Background()
			}

			result, err := jobAppsTool.HandlerFunc(ctx, mcp.CallToolRequest{})

			if tt.expectedError != nil {
				if err != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					require.NotNil(t, result)
					assert.True(t, result.IsError)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)

			if result.StructuredContent == nil {
				assert.Equal(t, 0, tt.expectedCount, "expected 0 results when StructuredContent is nil")
				return
			}

			applications, ok := result.StructuredContent.([]queries.GetAllJobApplicationsByUserIDRow)
			require.True(t, ok, "expected structured content to be []queries.GetAllJobApplicationsByUserIDRow, got %T", result.StructuredContent)
			assert.Len(t, applications, tt.expectedCount)
			if tt.expectedCount > 0 {
				for _, app := range applications {
					assert.NotEmpty(t, app.Company)
					assert.NotEmpty(t, app.Title)
					assert.NotEmpty(t, app.Status)
				}
			}
		})
	}
}
