package tool

import (
	"context"
	"errors"

	contextkey "github.com/Piszmog/pathwise/internal/context_key"
	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) NewJobApplicationsStatusHistoryTool() Tool {
	return Tool{
		Tool: mcp.NewTool(
			"job_applications_status_history",
			mcp.WithDescription("Get status history for job applications"),
			mcp.WithNumber("job_application_id", mcp.Description("ID of the job application to get status history for (optional - if not provided, returns history for all applications)")),
		),
		HandlerFunc: h.GetJobApplicationsStatusHistory,
	}
}

func (h *Handler) GetJobApplicationsStatusHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, ok := ctx.Value(contextkey.KeyUserID).(int64)
	if !ok {
		h.Logger.ErrorContext(ctx, "authentication failed - user ID not found in context", "tool", "job_applications_status_history")
		return mcp.NewToolResultError("failed to authenticate"), nil
	}

	var data []queries.JobApplicationStatusHistory
	var err error
	if jobAppID, exists := req.GetArguments()["job_application_id"]; exists {
		jobAppIDInt, ok := jobAppID.(float64)
		if !ok {
			h.Logger.ErrorContext(ctx, "invalid job_application_id parameter", "tool", "job_applications_status_history", "provided_value", jobAppID, "expected_type", "float64", "user_id", userID)
			return mcp.NewToolResultError("invalid job_application_id"), nil
		}

		data, err = h.Database.Queries().GetJobApplicationStatusHistoryByJobApplicationIDAndUserID(ctx, queries.GetJobApplicationStatusHistoryByJobApplicationIDAndUserIDParams{
			JobApplicationID: int64(jobAppIDInt),
			UserID:           userID,
		})
		if err != nil {
			h.Logger.ErrorContext(ctx, "failed to retrieve job application status history", "error", err, "user_id", userID, "job_application_id", int64(jobAppIDInt))
			return nil, errJobApplicationsStatusHistory
		}
	} else {
		data, err = h.Database.Queries().GetAllJobApplicationStatusHistoryByUserID(ctx, userID)
		if err != nil {
			h.Logger.ErrorContext(ctx, "failed to retrieve job applications status history", "error", err, "user_id", userID)
			return nil, errJobApplicationsStatusHistory
		}
	}
	return mcp.NewToolResultStructuredOnly(data), nil
}

var errJobApplicationsStatusHistory = errors.New("failed to retrieve job applications status history")
