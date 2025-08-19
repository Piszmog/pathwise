package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/logger"
	"github.com/Piszmog/pathwise/internal/version"
	"github.com/Piszmog/pathwise/mcp/server"
	"github.com/Piszmog/pathwise/mcp/tool"
)

func main() {
	l := logger.New(os.Getenv("LOG_LEVEL"), os.Getenv("LOG_OUTPUT"))

	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		l.Error("failed to create temp dir", "error", err)
		os.Exit(1)
	}
	defer func() {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			l.Error("failed to remove temp dir", "error", removeErr)
		}
	}()

	dbPath := filepath.Join(dir, "db-mcp.sqlite3")

	database, err := db.New(
		l,
		db.DatabaseOpts{
			URL:           dbPath,
			SyncURL:       os.Getenv("DB_PRIMARY_URL"),
			Token:         os.Getenv("DB_TOKEN_READONLY"),
			EncryptionKey: os.Getenv("ENC_KEY"),
			SyncInterval:  6 * time.Hour,
		},
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
		port = "8081"
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
