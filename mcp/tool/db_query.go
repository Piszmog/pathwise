package tool

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	contextkey "github.com/Piszmog/pathwise/internal/context_key"
	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) NewQueryDBTool() Tool {
	return Tool{
		Tool: mcp.NewTool(
			"db_query",
			mcp.WithDescription("List the tables available to be queried"),
			mcp.WithString(
				"query",
				mcp.Required(),
				mcp.Description("The SQLite query string."),
			),
			mcp.WithArray(
				"params",
				mcp.Required(),
				mcp.Description("The parameters to pass to the query. 'user_id' value will be injected by the MCP Server. Tables that do not have a 'user_id' column should be joined with another table that does have a 'user_id' column."),
			),
		),
		HandlerFunc: h.QueryDB,
	}
}

func (h *Handler) QueryDB(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, ok := ctx.Value(contextkey.KeyUserID).(int64)
	if !ok {
		return mcp.NewToolResultError("failed to authenticate"), nil
	}
	var args struct {
		Query  string `json:"query"`
		Params []any  `json:"params"`
	}

	if err := req.BindArguments(&args); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	query := args.Query
	params := append([]any{userID}, args.Params...)

	if !strings.Contains(strings.ToUpper(query), "WHERE USER_ID = ?") {
		return mcp.NewToolResultError("queries must include 'WHERE user_id = ?' clause"), nil
	}

	if !isSelectQuery(query) {
		return mcp.NewToolResultError("only SELECT queries are allowed"), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := h.Database.DB().QueryContext(ctx, query, params...)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("query failed: %v", err)), nil
	}
	defer func() {
		if err := rows.Close(); err != nil {
			h.Logger.ErrorContext(ctx, "failed to close rows", "error", err, "query", query)
		}
	}()

	results, err := rowsToJSON(rows)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to process results: %v", err)), nil
	}

	resultData := map[string]any{
		"query":   query,
		"results": results,
		"count":   len(results),
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func rowsToJSON(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]any

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

func isSelectQuery(query string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	return strings.HasPrefix(trimmed, "SELECT")
}
