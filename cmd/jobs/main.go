package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/logger"
	"github.com/Piszmog/pathwise/jobs/hn"
	"github.com/Piszmog/pathwise/jobs/llm"
)

func main() {
	l := logger.New(os.Getenv("LOG_LEVEL"), os.Getenv("LOG_OUTPUT"))

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

	if _, err = database.DB().Exec("PRAGMA foreign_keys = ON;"); err != nil {
		l.Error("failed to enable foreign keys", "error", err)
		return
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	scraper := hn.NewScraper(l, database, httpClient)

	commentIDsChan := make(chan int64, 1000)

	processor := hn.NewProcessor(l, database, llmClient)
	go processor.Run(ctx, commentIDsChan)

	go func() {
		l.DebugContext(ctx, "running scraper")
		if scrapeErr := scraper.Run(ctx, commentIDsChan); scrapeErr != nil {
			l.Error("failed to scrape", "error", scrapeErr)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go func() {
				l.DebugContext(ctx, "running scraper")
				if err := scraper.Run(ctx, commentIDsChan); err != nil {
					l.Error("failed to scrape", "error", err)
				}
			}()
		case <-sigChan:
			l.Info("shutting down...")
			close(commentIDsChan)
			cancel()
			return
		}
	}
}
