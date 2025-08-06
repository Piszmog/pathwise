package tool

import (
	"context"
	"errors"

	contextkey "github.com/Piszmog/pathwise/internal/context_key"
	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) NewJobApplicationsTool() Tool {
	return Tool{
		Tool: mcp.NewTool(
			"job_applications",
			mcp.WithDescription("Information about your job applications"),
		),
		HandlerFunc: h.GetJobApplications,
	}
}

func (h *Handler) GetJobApplications(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, ok := ctx.Value(contextkey.KeyUserID).(int64)
	if !ok {
		return mcp.NewToolResultError("failed to authenticate"), nil
	}

	data, err := h.Database.Queries().GetAllJobApplicationsByUserID(ctx, userID)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to retrieve job applications", "error", err, "user_id", userID)
		return nil, errors.New("failed to retrieve job applications")
	}
	return mcp.NewToolResultStructuredOnly(data), nil
}
