package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/jobs/hn"
	"github.com/Piszmog/pathwise/internal/jobs/llm"
	"github.com/Piszmog/pathwise/internal/jobs/server/router"
	"github.com/Piszmog/pathwise/internal/logger"
	"github.com/Piszmog/pathwise/internal/server"
)

func main() {
	l := logger.New(os.Getenv("LOG_LEVEL"), os.Getenv("LOG_OUTPUT"))
	l.Info("starting jobs app")

	geminiToken := os.Getenv("GEMINI_API_KEY")
	if geminiToken == "" {
		l.Error("missing Gemini token")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	llmClient, err := llm.NewGeminiClient(ctx, geminiToken)
	if err != nil {
		l.Error("failed to create Gemini client", "error", err)
		return
	}
	defer func() {
		llmErr := llmClient.Close()
		if llmErr != nil {
			l.Error("failed to close the Gemini client", "error", llmErr)
		}
	}()

	dbURL := os.Getenv("DB_URL")
	var tempDir string
	if dbURL == "" {
		dir, err := os.MkdirTemp("", "libsql-*")
		if err != nil {
			l.Error("failed to create temp dir", "error", err)
			return
		}
		tempDir = dir
		defer func() {
			if removeErr := os.RemoveAll(tempDir); removeErr != nil {
				l.Error("failed to remove temp dir", "error", removeErr)
			}
		}()
		dbURL = filepath.Join(tempDir, "db-jobs.sqlite3")
	}

	database, err := db.New(
		l,
		db.DatabaseOpts{
			URL:           dbURL,
			SyncURL:       os.Getenv("DB_PRIMARY_URL"),
			Token:         os.Getenv("DB_TOKEN"),
			EncryptionKey: os.Getenv("ENC_KEY"),
			SyncInterval:  12 * time.Hour,
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

	if _, err = database.DB().Exec("PRAGMA foreign_keys = ON;"); err != nil {
		l.Error("failed to enable foreign keys", "error", err)
		return
	}

	hnRunner := hn.NewRunner(l, database, llmClient)
	hnRunner.Run(ctx)
	defer func() {
		_ = hnRunner.Close()
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	r := router.New(l, database)
	server.New(l, ":"+port, server.WithHandler(r)).StartAndWait()
	cancel()
}
