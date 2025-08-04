package middleware

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log/slog"

	contextkey "github.com/Piszmog/pathwise/internal/context_key"
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
		apiKey := request.Header.Get("X-API-Key")
		if apiKey == "" {
			m.Logger.WarnContext(ctx, "missing API key")
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Authentication required: missing API key",
					},
				},
			}, nil
		}

		keyHash := m.hashAPIKey(apiKey)

		userID, err := m.Database.Queries().GetMcpAPIKeyByHash(ctx, keyHash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				m.Logger.WarnContext(ctx, "invalid API key")
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						mcp.TextContent{
							Type: "text",
							Text: "Authentication failed: invalid API key",
						},
					},
				}, nil
			}
			m.Logger.ErrorContext(ctx, "failed to get API key", "err", err)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Internal server error",
					},
				},
			}, nil
		}

		ctx = context.WithValue(ctx, contextkey.KeyUserID, userID)
		m.Logger.DebugContext(ctx, "authenticated user", "user_id", userID)

		return next(ctx, request)
	}
}

func (m *AuthMiddleware) hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}
