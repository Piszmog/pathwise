package main

import (
	"context"
	"embed"
	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/handlers"
	"github.com/Piszmog/pathwise/logger"
	"github.com/Piszmog/pathwise/middleware"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

//go:embed assets/*
var assets embed.FS

func main() {
	l := logger.New(os.Getenv("LOG_LEVEL"))

	database, err := db.New(db.DatabaseTypeFile, db.DatabaseOpts{URL: "./db.sqlite3"})
	if err != nil {
		l.Error("failed to create database", "error", err)
		return
	}
	defer database.Close()

	if err = db.Init(database); err != nil {
		l.Error("failed to initialize database", "error", err)
		return
	}

	handler := &handlers.Handler{
		Logger:                           l,
		JobApplicationStore:              &store.JobApplicationStore{Database: database},
		JobApplicationNoteStore:          &store.JobApplicationNoteStore{Database: database},
		JobApplicationStatusHistoryStore: &store.JobApplicationStatusHistoryStore{Database: database},
		StatsStore:                       &store.StatsStore{Database: database},
	}

	r := mux.NewRouter()
	r.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assets)))
	r.HandleFunc("/", handler.Main).Methods(http.MethodGet)
	r.HandleFunc("/jobs", handler.AddJob).Methods(http.MethodPost)
	r.HandleFunc("/jobs", handler.GetJobs).Methods(http.MethodGet)
	r.HandleFunc("/jobs/{id}", handler.JobDetails).Methods(http.MethodGet)
	r.HandleFunc("/jobs/{id}", handler.UpdateJob).Methods(http.MethodPatch)
	r.HandleFunc("/jobs/{id}/notes", handler.AddNote).Methods(http.MethodPost)

	loggingMiddleware := middleware.LoggingMiddleware{Logger: l}
	r.Use(loggingMiddleware.Middleware)

	srv := &http.Server{
		Handler: r,
		Addr:    ":8080",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		l.Info("starting server", "port", "8080")
		if err = srv.ListenAndServe(); err != nil {
			l.Warn("failed to start server", "error", err)
		}
	}()

	gracefulShutdown(srv, l)
}

func gracefulShutdown(srv *http.Server, l *slog.Logger) {
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	l.Info("shutting down")
	os.Exit(0)
}
