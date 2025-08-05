package main

import (
	"os"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/logger"
	"github.com/Piszmog/pathwise/internal/version"
	"github.com/Piszmog/pathwise/mcp/server"
	"github.com/Piszmog/pathwise/mcp/tool"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	l := logger.New(os.Getenv("LOG_LEVEL"), os.Getenv("LOG_OUTPUT"))

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "./db.sqlite3"
	}

	database, err := db.New(
		l,
		db.DatabaseOpts{URL: dbURL, Token: os.Getenv("DB_TOKEN")},
	)
	if err != nil {
		l.Error("failed to create database", "error", err)
		return
	}
	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			l.Error("failed to close database", "error", closeErr)
		}
	}()

	v := os.Getenv("VERSION")
	if v != "" {
		version.Value = v
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	toolHandlers := tool.Handler{Logger: l, Database: database}

	srv := server.New(
		"Pathwise MCP Server",
		":"+port,
		l,
		database,
		server.AddTool(
			"list_tables",
			"List the tables available to be queries",
			toolHandlers.ListTables,
		),
		server.AddTool(
			"db_query",
			"List the tables available to be queries",
			toolHandlers.QueryDB,
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
	)

	if err = srv.Start(); err != nil {
		l.Error("failed to run the MCP server", "error", err)
		return
	}
}
