package tool

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) NewListTablesTool() Tool {
	return Tool{
		Tool: mcp.NewTool(
			"list_tables",
			mcp.WithDescription("List the tables available to be queried"),
		),
		HandlerFunc: h.ListTables,
	}
}

type tableInfo struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Columns     []columnInfo `json:"columns"`
	HasUserID   bool         `json:"hasUserId"`
}

type columnInfo struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	NotNull      bool    `json:"notNull"`
	DefaultValue *string `json:"defaultValue,omitempty"`
	PrimaryKey   bool    `json:"primaryKey"`
}

func (h *Handler) ListTables(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tables := h.getAvailableTables(ctx)

	response := map[string]any{
		"tables":      tables,
		"count":       len(tables),
		"description": "Tables accessible to your account",
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (h *Handler) getAvailableTables(ctx context.Context) []tableInfo {
	userTables := map[string]string{
		"job_applications":                 "Job applications you've submitted",
		"job_application_notes":            "Notes on your job applications",
		"job_application_stat":             "Statistics about your job applications",
		"job_application_status_histories": "Status change history for your applications",
	}

	tables := make([]tableInfo, 0, len(userTables))
	for tableName, description := range userTables {
		columns, err := h.getTableColumns(ctx, tableName)
		if err != nil {
			continue
		}

		hasUserID := h.tableHasUserID(columns)

		tables = append(tables, tableInfo{
			Name:        tableName,
			Description: description,
			Columns:     columns,
			HasUserID:   hasUserID,
		})
	}

	return tables
}

func (h *Handler) getTableColumns(ctx context.Context, tableName string) ([]columnInfo, error) {
	// Whitelist validation for tableName
	allowedTables := map[string]struct{}{
		"job_applications":                 {},
		"job_application_notes":            {},
		"job_application_stat":             {},
		"job_application_status_histories": {},
	}
	if _, ok := allowedTables[tableName]; !ok {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	rows, err := h.Database.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			h.Logger.ErrorContext(ctx, "failed to close rows", "error", closeErr, "table", tableName)
		}
	}()

	var columns []columnInfo
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return nil, err
		}

		col := columnInfo{
			Name:       name,
			Type:       dataType,
			NotNull:    notNull == 1,
			PrimaryKey: pk == 1,
		}

		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

func (h *Handler) tableHasUserID(columns []columnInfo) bool {
	for _, col := range columns {
		if col.Name == "user_id" {
			return true
		}
	}
	return false
}
