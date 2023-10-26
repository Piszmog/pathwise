package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/handlers"
	"github.com/Piszmog/pathwise/types"
	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

//go:embed assets/*
var assets embed.FS

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assets)))

	m := components.Main(
		[]types.JobApplication{
			{
				ID:        1,
				Company:   "Company 1",
				Title:     "Title 1",
				Status:    types.JobApplicationStatusApplied,
				AppliedAt: time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        2,
				Company:   "Company 2",
				Title:     "Title 2",
				Status:    types.JobApplicationStatusApplied,
				AppliedAt: time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		types.StatsOpts{
			TotalApplications:           "2",
			TotalCompanies:              "2",
			AverageTimeToHearBackInDays: "2",
			TotalInterviewingPercentage: "4",
			TotalRejectionsPercentage:   "44",
		},
	)
	r.Handle("/", templ.Handler(m))
	r.HandleFunc("/jobs/{id}", handlers.GetJob).Methods(http.MethodGet)

	srv := &http.Server{
		Handler: r,
		Addr:    ":8080",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		fmt.Println("Listening on port 8080")
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	gracefulShutdown(srv)
}

func gracefulShutdown(srv *http.Server) {
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
	log.Println("shutting down")
	os.Exit(0)
}
