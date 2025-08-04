package tool

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

type tableInfo struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Columns     []columnInfo `json:"columns"`
	HasUserID   bool         `json:"has_user_id"`
}

type columnInfo struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	NotNull      bool    `json:"not_null"`
	DefaultValue *string `json:"default_value,omitempty"`
	PrimaryKey   bool    `json:"primary_key"`
}

func (h *Handler) ListTables(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tables, err := h.getAvailableTables(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get tables: %v", err)), nil
	}

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

func (h *Handler) getAvailableTables(ctx context.Context) ([]tableInfo, error) {
	userTables := map[string]string{
		"job_applications":                 "Job applications you've submitted",
		"job_application_notes":            "Notes on your job applications",
		"job_application_stat":             "Statistics about your job applications",
		"job_application_status_histories": "Status change history for your applications",
	}

	var tables []tableInfo
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

	return tables, nil
}

func (h *Handler) getTableColumns(ctx context.Context, tableName string) ([]columnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	rows, err := h.Database.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
