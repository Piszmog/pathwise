package router

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/server/handler"
	"github.com/Piszmog/pathwise/server/middleware"
	"github.com/gorilla/mux"
)

func New(logger *slog.Logger, database db.Database, assets embed.FS, sessionStore *store.SessionStore) http.Handler {
	h := &handler.Handler{
		Logger:                           logger,
		JobApplicationStore:              &store.JobApplicationStore{Database: database},
		JobApplicationNoteStore:          &store.JobApplicationNoteStore{Database: database},
		JobApplicationStatusHistoryStore: &store.JobApplicationStatusHistoryStore{Database: database},
		StatsStore:                       &store.StatsStore{Database: database},
		UserStore:                        &store.UserStore{Database: database},
		SessionsStore:                    sessionStore,
	}

	r := mux.NewRouter()
	loggingMiddleware := middleware.LoggingMiddleware{Logger: logger}
	r.Use(loggingMiddleware.Middleware)
	// TODO: CORS

	r.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assets)))
	r.HandleFunc("/signup", h.Signup).Methods(http.MethodGet)
	r.HandleFunc("/signup", h.Register).Methods(http.MethodPost)
	r.HandleFunc("/signin", h.Signin).Methods(http.MethodGet)
	r.HandleFunc("/signin", h.Authenticate).Methods(http.MethodPost)

	protected := r.NewRoute().Subrouter()
	authMiddleware := middleware.AuthMiddleware{
		Logger:       logger,
		SessionStore: sessionStore,
	}
	protected.Use(authMiddleware.Middleware)

	protected.HandleFunc("/", h.Main).Methods(http.MethodGet)
	protected.HandleFunc("/jobs", h.AddJob).Methods(http.MethodPost)
	protected.HandleFunc("/jobs", h.GetJobs).Methods(http.MethodGet)
	protected.HandleFunc("/jobs/{id}", h.JobDetails).Methods(http.MethodGet)
	protected.HandleFunc("/jobs/{id}", h.UpdateJob).Methods(http.MethodPatch)
	protected.HandleFunc("/jobs/{id}/notes", h.AddNote).Methods(http.MethodPost)

	return r
}
