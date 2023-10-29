package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/handlers"
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
	database, err := db.New(db.DatabaseTypeFile, db.DatabaseOpts{URL: "./db.sqlite3"})
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err = db.Init(database); err != nil {
		log.Fatal(err)
	}

	handler := &handlers.Handler{
		JobApplicationStore:              &store.JobApplicationStore{Database: database},
		JobApplicationNoteStore:          &store.JobApplicationNoteStore{Database: database},
		JobApplicationStatusHistoryStore: &store.JobApplicationStatusHistoryStore{Database: database},
		StatsStore:                       &store.StatsStore{Database: database},
	}

	r := mux.NewRouter()
	r.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assets)))
	r.HandleFunc("/", handler.Jobs).Methods(http.MethodGet)
	r.HandleFunc("/jobs", handler.AddJob).Methods(http.MethodPost)
	r.HandleFunc("/jobs/{id}", handler.JobDetails).Methods(http.MethodGet)
	r.HandleFunc("/jobs/{id}", handler.UpdateJob).Methods(http.MethodPatch)
	r.HandleFunc("/jobs/{id}/notes", handler.AddNote).Methods(http.MethodPost)

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
