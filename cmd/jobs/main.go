package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/logger"
	"github.com/Piszmog/pathwise/jobs/hn"
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

	if _, err = database.DB().Exec("PRAGMA foreign_keys = ON;"); err != nil {
		l.Error("failed to enable foreign keys", "error", err)
		return
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	scraper := hn.NewScraper(l, database, httpClient)

	commentIDsChan := make(chan int64, 1000)
	err = scraper.Run(context.Background(), commentIDsChan)
	if err != nil {
		l.Error("failed to scrape", "error", err)
		return
	}
}
