package tool

import (
	"context"
	"errors"

	contextkey "github.com/Piszmog/pathwise/internal/context_key"
	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) NewJobApplicationsNotesTool() Tool {
	return Tool{
		Tool: mcp.NewTool(
			"job_applications_notes",
			mcp.WithDescription("Get notes for job applications"),
			mcp.WithNumber("job_application_id", mcp.Description("ID of the job application to get notes for (optional - if not provided, returns notes for all applications)")),
		),
		HandlerFunc: h.GetJobApplicationsNotes,
	}
}

func (h *Handler) GetJobApplicationsNotes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, ok := ctx.Value(contextkey.KeyUserID).(int64)
	if !ok {
		return mcp.NewToolResultError("failed to authenticate"), nil
	}

	if jobAppID, exists := req.GetArguments()["job_application_id"]; exists {
		jobAppIDInt, ok := jobAppID.(int64)
		if !ok {
			return mcp.NewToolResultError("invalid job_application_id"), nil
		}

		data, err := h.Database.Queries().GetJobApplicationNotesByJobApplicationIDAndUserID(ctx, queries.GetJobApplicationNotesByJobApplicationIDAndUserIDParams{
			JobApplicationID: jobAppIDInt,
			UserID:           userID,
		})
		if err != nil {
			h.Logger.ErrorContext(ctx, "failed to retrieve job application notes", "error", err, "user_id", userID, "job_application_id", jobAppIDInt)
			return nil, errJobApplicationsNotes
		}
		return mcp.NewToolResultStructuredOnly(data), nil
	} else {
		data, err := h.Database.Queries().GetAllJobApplicationNotesByUserID(ctx, userID)
		if err != nil {
			h.Logger.ErrorContext(ctx, "failed to retrieve job applications notes", "error", err, "user_id", userID)
			return nil, errJobApplicationsNotes
		}
		return mcp.NewToolResultStructuredOnly(data), nil
	}
}

var errJobApplicationsNotes = errors.New("failed to retrieve job applications notes")
