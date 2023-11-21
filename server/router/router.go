package router

import (
	"embed"
	"log/slog"
	"net/http"
	"os"

	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/server/handler"
	"github.com/Piszmog/pathwise/server/middleware"
	"github.com/gorilla/mux"
)

func New(logger *slog.Logger, database db.Database, assets embed.FS, sessionStore *store.SessionStore) http.Handler {
	version := getVersion()
	h := &handler.Handler{
		Version:                          version,
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

	cache := middleware.CacheControlMiddleware{Version: version}
	r.PathPrefix("/assets/").Handler(cache.Middleware(http.FileServer(http.FS(assets))))
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
	protected.HandleFunc("/signout", h.Signout).Methods(http.MethodGet)
	protected.HandleFunc("/settings", h.Settings).Methods(http.MethodGet)
	protected.HandleFunc("/settings/changePassword", h.ChangePassword).Methods(http.MethodPost)
	protected.HandleFunc("/settings/logoutSessions", h.LogoutSessions).Methods(http.MethodPost)
	protected.HandleFunc("/settings/deleteAccount", h.DeleteAccount).Methods(http.MethodPost)

	return r
}

func getVersion() string {
	version := os.Getenv("VERSION")
	if version == "" {
		version = "dev"
	}
	return version
}
