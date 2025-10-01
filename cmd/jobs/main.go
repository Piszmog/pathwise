package main

import (
	"context"
	"log/slog"
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

	scraper := hn.NewScraper(l, database, &http.Client{Timeout: 10 * time.Second})
	processor := hn.NewProcessor(l, database, llmClient)
	commentIDsChan := make(chan int64, 1000)

	go processor.Run(ctx, commentIDsChan)
	go startCommentProcessor(ctx, l, database, commentIDsChan)
	go startScraper(ctx, l, scraper, commentIDsChan)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	l.Info("shutting down...")
	cancel()
	close(commentIDsChan)
	time.Sleep(100 * time.Millisecond)
	close(commentIDsChan)
}

func startCommentProcessor(ctx context.Context, logger *slog.Logger, database db.Database, commentIDsChan chan<- int64) {
	processQueuedComments(ctx, logger, database, []string{"queued", "in_progress", "failed"}, commentIDsChan)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			processQueuedComments(ctx, logger, database, []string{"failed"}, commentIDsChan)
		case <-ctx.Done():
			return
		}
	}
}

func processQueuedComments(ctx context.Context, logger *slog.Logger, database db.Database, status []string, commentIDsChan chan<- int64) {
	logger.DebugContext(ctx, "getting un-completed comments")
	ids, err := database.Queries().GetQueuedHNComments(ctx, status)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get queued comments", "error", err)
		return
	}
	for _, id := range ids {
		commentIDsChan <- id
	}
}

func startScraper(ctx context.Context, logger *slog.Logger, scraper *hn.Scraper, commentIDsChan chan<- int64) {
	runScraper(ctx, logger, scraper, commentIDsChan)

	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runScraper(ctx, logger, scraper, commentIDsChan)
		case <-ctx.Done():
			return
		}
	}
}

func runScraper(ctx context.Context, logger *slog.Logger, scraper *hn.Scraper, commentIDsChan chan<- int64) {
	logger.DebugContext(ctx, "running scraper")
	if err := scraper.Run(ctx, commentIDsChan); err != nil {
		logger.ErrorContext(ctx, "failed to scrape", "error", err)
	}
}
