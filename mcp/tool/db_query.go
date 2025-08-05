package tool

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	contextkey "github.com/Piszmog/pathwise/internal/context_key"
	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) NewQueryDBTool() Tool {
	return Tool{
		Tool: mcp.NewTool(
			"db_query",
			mcp.WithDescription("Perform a secure database query"),
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

	// Enhanced security validation BEFORE parameter injection
	if err := validateSecureQuery(query); err != nil {
		h.Logger.WarnContext(ctx, "rejected insecure query",
			"error", err,
			"query", query,
			"user_id", userID)
		return mcp.NewToolResultError(fmt.Sprintf("security validation failed: %v", err)), nil
	}

	// Inject user_id as first parameter after validation
	params := append([]any{userID}, args.Params...)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := h.Database.DB().QueryContext(ctx, query, params...)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("query failed: %v", err)), nil
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			h.Logger.ErrorContext(ctx, "failed to close rows", "error", closeErr, "query", query)
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

var (
	ErrCommentsNotAllowed   = errors.New("comments not allowed in queries")
	ErrSubqueriesNotAllowed = errors.New("subqueries not allowed")
	ErrInvalidWhereClause   = errors.New("queries must start with 'WHERE user_id = ?' as the first WHERE clause")
	ErrForbiddenKeyword     = errors.New("forbidden keyword")
)

func validateSecureQuery(query string) error {
	// 1. Block comments
	if strings.Contains(query, "/*") || strings.Contains(query, "--") {
		return ErrCommentsNotAllowed
	}

	// 2. Block dangerous keywords
	forbidden := []string{"UNION", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "EXEC", "PRAGMA", "ATTACH", "DETACH"}
	upperQuery := strings.ToUpper(query)
	for _, keyword := range forbidden {
		if strings.Contains(upperQuery, keyword) {
			return fmt.Errorf("%w: %s", ErrForbiddenKeyword, keyword)
		}
	}

	// 3. Block subqueries
	if strings.Contains(upperQuery, "(SELECT") {
		return ErrSubqueriesNotAllowed
	}

	// 4. Ensure the query contains 'WHERE USER_ID = ?' as the first WHERE clause
	// Note: user_id parameter is automatically injected as the first parameter
	re := regexp.MustCompile(`(?i)^\s*SELECT\s+.*\s+FROM\s+.*\s+WHERE\s+USER_ID\s*=\s*\?\s*(?:AND|ORDER BY|GROUP BY|LIMIT|$)`)
	if !re.MatchString(query) {
		return ErrInvalidWhereClause
	}

	return nil
}
