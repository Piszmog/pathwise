package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/jobs/hn"
	"github.com/Piszmog/pathwise/internal/jobs/llm"
	"github.com/Piszmog/pathwise/internal/logger"
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
	dbPath := filepath.Join(dir, "db-jobs.sqlite3")

	database, err := db.New(
		l,
		db.DatabaseOpts{
			URL:           dbPath,
			SyncURL:       os.Getenv("DB_PRIMARY_URL"),
			Token:         os.Getenv("DB_TOKEN_READONLY"),
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	l.Info("shutting down...")
	cancel()
	time.Sleep(100 * time.Millisecond)
}
