package router

import (
	"embed"
	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/server/handler"
	"github.com/Piszmog/pathwise/server/middleware"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
)

//go:embed assets/*
var assets embed.FS

func New(logger *slog.Logger, database db.Database) http.Handler {
	h := &handler.Handler{
		Logger:                           logger,
		JobApplicationStore:              &store.JobApplicationStore{Database: database},
		JobApplicationNoteStore:          &store.JobApplicationNoteStore{Database: database},
		JobApplicationStatusHistoryStore: &store.JobApplicationStatusHistoryStore{Database: database},
		StatsStore:                       &store.StatsStore{Database: database},
	}

	r := mux.NewRouter()
	r.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assets)))
	r.HandleFunc("/", h.Main).Methods(http.MethodGet)
	r.HandleFunc("/jobs", h.AddJob).Methods(http.MethodPost)
	r.HandleFunc("/jobs", h.GetJobs).Methods(http.MethodGet)
	r.HandleFunc("/jobs/{id}", h.JobDetails).Methods(http.MethodGet)
	r.HandleFunc("/jobs/{id}", h.UpdateJob).Methods(http.MethodPatch)
	r.HandleFunc("/jobs/{id}/notes", h.AddNote).Methods(http.MethodPost)

	loggingMiddleware := middleware.LoggingMiddleware{Logger: logger}
	r.Use(loggingMiddleware.Middleware)

	return r
}
