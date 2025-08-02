package main

import (
	"os"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/logger"
	"github.com/Piszmog/pathwise/internal/version"
	"github.com/Piszmog/pathwise/mcp/server"
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

	srv := server.New(
		"Pathwise MCP Server",
		":"+port,
		l,
	)

	if err = srv.Start(); err != nil {
		l.Error("failed to run the MCP server", "error", err)
		os.Exit(1)
	}
}
