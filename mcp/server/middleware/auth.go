package middleware

import (
	"context"
	"log/slog"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type AuthMiddleware struct {
	Logger   *slog.Logger
	Database db.Database
}

func (m *AuthMiddleware) Handle(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
		return next(ctx, request)
	}
}
