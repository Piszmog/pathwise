package main

import (
	"os"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/logger"
	"github.com/Piszmog/pathwise/internal/version"
	"github.com/Piszmog/pathwise/mcp/server"
	"github.com/Piszmog/pathwise/mcp/tool"
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
		server.AddTool(toolHandlers.NewJobApplicationsTool()),
		server.AddTool(toolHandlers.NewJobApplicationsStatusHistoryTool()),
		server.AddTool(toolHandlers.NewJobApplicationsNotesTool()),
	)

	if err = srv.Start(); err != nil {
		l.Error("failed to run the MCP server", "error", err)
		return
	}
}
