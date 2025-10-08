package hn

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/jobs/llm"
)

type Runner struct {
	scraper        *Scraper
	processor      *Processor
	database       db.Database
	logger         *slog.Logger
	commentIDsChan chan int64
}

func NewRunner(logger *slog.Logger, database db.Database, llmClient *llm.GeminiClient) *Runner {
	return &Runner{
		scraper:        NewScraper(logger, database, &http.Client{Timeout: 10 * time.Second}),
		processor:      NewProcessor(logger, database, llmClient),
		database:       database,
		logger:         logger,
		commentIDsChan: make(chan int64, 1000),
	}
}

func (r *Runner) Run(ctx context.Context) {
	go r.processor.Run(ctx, r.commentIDsChan)
	go r.startCommentProcessor(ctx)
	go r.startScraper(ctx)
}

func (r *Runner) Close() error {
	close(r.commentIDsChan)
	return nil
}

func (r *Runner) startCommentProcessor(ctx context.Context) {
	r.processQueuedComments(ctx, []string{"queued", "in_progress", "failed"})

	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.processQueuedComments(ctx, []string{"failed"})
		case <-ctx.Done():
			return
		}
	}
}

func (r *Runner) processQueuedComments(ctx context.Context, status []string) {
	r.logger.DebugContext(ctx, "getting un-completed comments")
	ids, err := r.database.Queries().GetQueuedHNComments(ctx, status)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to get queued comments", "error", err)
		return
	}
	for _, id := range ids {
		r.commentIDsChan <- id
	}
}

func (r *Runner) startScraper(ctx context.Context) {
	r.runScraper(ctx)

	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.runScraper(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (r *Runner) runScraper(ctx context.Context) {
	r.logger.DebugContext(ctx, "running scraper")
	if err := r.scraper.Run(ctx, r.commentIDsChan); err != nil {
		r.logger.ErrorContext(ctx, "failed to scrape", "error", err)
	}
}
