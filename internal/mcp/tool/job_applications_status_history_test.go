//go:build integration

package tool_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	contextkey "github.com/Piszmog/pathwise/internal/context_key"
	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/Piszmog/pathwise/internal/mcp/tool"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/tursodatabase/go-libsql"
)

var errInvalidJobAppIDStatusHistory = errors.New("invalid job_application_id for status history")

type statusHistoryTestCase struct {
	name             string
	userID           int64
	jobApplicationID interface{}
	setupData        func(t *testing.T, db *sql.DB, userID int64) map[string]int64
	expectedCount    int
	expectedError    error
}

func TestJobApplicationsStatusHistoryTool(t *testing.T) {
	tests := []statusHistoryTestCase{
		{
			name:             "unauthenticated user",
			userID:           0,
			jobApplicationID: nil,
			setupData:        func(t *testing.T, db *sql.DB, userID int64) map[string]int64 { return nil },
			expectedCount:    0,
			expectedError:    errAuthenticationFailed,
		},
		{
			name:             "no additional status history for any applications",
			userID:           1,
			jobApplicationID: nil,
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				jobID := insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
				return map[string]int64{"jobID": jobID}
			},
			expectedCount: 1,
			expectedError: nil,
		},
		{
			name:             "single additional status history across applications",
			userID:           1,
			jobApplicationID: nil,
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				jobID := insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
				insertJobApplicationStatusHistory(t, db, jobID, "interviewing")
				return map[string]int64{"jobID": jobID}
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name:             "multiple status history across applications",
			userID:           1,
			jobApplicationID: nil,
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				jobID1 := insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
				jobID2 := insertJobApplication(t, db, userID, "Company B", "Developer", "applied")
				insertJobApplicationStatusHistory(t, db, jobID1, "interviewing")
				insertJobApplicationStatusHistory(t, db, jobID1, "offered")
				insertJobApplicationStatusHistory(t, db, jobID2, "rejected")
				return map[string]int64{"jobID1": jobID1, "jobID2": jobID2}
			},
			expectedCount: 5,
			expectedError: nil,
		},
		{
			name:             "user isolation - only sees own status history",
			userID:           1,
			jobApplicationID: nil,
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				createTestUser(t, db, 2)
				otherJobID := insertJobApplication(t, db, 2, "Other Company", "Other Job", "applied")
				insertJobApplicationStatusHistory(t, db, otherJobID, "interviewing")

				myJobID := insertJobApplication(t, db, userID, "My Company", "My Job", "applied")
				insertJobApplicationStatusHistory(t, db, myJobID, "interviewing")
				return map[string]int64{"myJobID": myJobID, "otherJobID": otherJobID}
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name:             "no additional status history for specific application",
			userID:           1,
			jobApplicationID: "jobID",
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				jobID := insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
				return map[string]int64{"jobID": jobID}
			},
			expectedCount: 1,
			expectedError: nil,
		},
		{
			name:             "single additional status history for specific application",
			userID:           1,
			jobApplicationID: "jobID",
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				jobID := insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
				insertJobApplicationStatusHistory(t, db, jobID, "interviewing")
				return map[string]int64{"jobID": jobID}
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name:             "multiple status history for specific application",
			userID:           1,
			jobApplicationID: "jobID1",
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				jobID1 := insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
				jobID2 := insertJobApplication(t, db, userID, "Company B", "Developer", "applied")
				insertJobApplicationStatusHistory(t, db, jobID1, "interviewing")
				insertJobApplicationStatusHistory(t, db, jobID1, "offered")
				insertJobApplicationStatusHistory(t, db, jobID2, "rejected")
				return map[string]int64{"jobID1": jobID1, "jobID2": jobID2}
			},
			expectedCount: 3,
			expectedError: nil,
		},
		{
			name:             "invalid job_application_id type - string",
			userID:           1,
			jobApplicationID: "invalid_string",
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				jobID := insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
				return map[string]int64{"jobID": jobID}
			},
			expectedCount: 0,
			expectedError: errInvalidJobAppIDStatusHistory,
		},
		{
			name:             "invalid job_application_id type - int64",
			userID:           1,
			jobApplicationID: int64(123),
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				jobID := insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
				return map[string]int64{"jobID": jobID}
			},
			expectedCount: 0,
			expectedError: errInvalidJobAppIDStatusHistory,
		},
		{
			name:             "nonexistent job_application_id",
			userID:           1,
			jobApplicationID: float64(99999),
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				jobID := insertJobApplication(t, db, userID, "Company A", "Engineer", "applied")
				return map[string]int64{"jobID": jobID}
			},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name:             "job_application_id belongs to other user",
			userID:           1,
			jobApplicationID: "otherJobID",
			setupData: func(t *testing.T, db *sql.DB, userID int64) map[string]int64 {
				createTestUser(t, db, 2)
				otherJobID := insertJobApplication(t, db, 2, "Other Company", "Other Job", "applied")
				insertJobApplicationStatusHistory(t, db, otherJobID, "interviewing")

				myJobID := insertJobApplication(t, db, userID, "My Company", "My Job", "applied")
				return map[string]int64{"myJobID": myJobID, "otherJobID": otherJobID}
			},
			expectedCount: 0,
			expectedError: nil,
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

			var jobIDs map[string]int64
			if tt.userID > 0 {
				createTestUser(t, database.DB(), tt.userID)
				jobIDs = tt.setupData(t, database.DB(), tt.userID)
			}

			statusHistoryTool := handler.NewJobApplicationsStatusHistoryTool()

			var ctx context.Context
			if tt.userID > 0 {
				ctx = context.WithValue(context.Background(), contextkey.KeyUserID, tt.userID)
			} else {
				ctx = context.Background()
			}

			req := mcp.CallToolRequest{}
			if tt.jobApplicationID != nil {
				var jobAppID interface{}
				if jobIDKey, ok := tt.jobApplicationID.(string); ok && jobIDs != nil {
					if actualJobID, exists := jobIDs[jobIDKey]; exists {
						jobAppID = float64(actualJobID)
					} else {
						jobAppID = tt.jobApplicationID
					}
				} else if intID, ok := tt.jobApplicationID.(int64); ok {
					jobAppID = intID
				} else {
					jobAppID = tt.jobApplicationID
				}

				req.Params.Arguments = map[string]interface{}{
					"job_application_id": jobAppID,
				}
			}
			result, err := statusHistoryTool.HandlerFunc(ctx, req)

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

			statusHistory, ok := result.StructuredContent.([]queries.JobApplicationStatusHistory)
			require.True(t, ok, "expected structured content to be []queries.JobApplicationStatusHistory, got %T", result.StructuredContent)
			assert.Len(t, statusHistory, tt.expectedCount)

			if tt.expectedCount > 0 {
				for _, history := range statusHistory {
					assert.NotEmpty(t, history.Status)
					assert.NotZero(t, history.JobApplicationID)
					assert.NotZero(t, history.CreatedAt)
				}
			}
		})
	}
}
