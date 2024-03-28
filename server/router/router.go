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

	router := http.NewServeMux()
	cache := middleware.CacheControlMiddleware{Version: version}
	router.Handle(http.MethodGet+" /assets/", cache.Middleware(http.FileServer(http.FS(assets))))
	router.HandleFunc(http.MethodGet+" /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/assets/img/favicon.ico", http.StatusSeeOther)
	})
	router.HandleFunc(http.MethodGet+" /signup", h.Signup)
	router.HandleFunc(http.MethodPost+" /signup", h.Register)
	router.HandleFunc(http.MethodGet+" /signin", h.Signin)
	router.HandleFunc(http.MethodPost+" /signin", h.Authenticate)

	protected := http.NewServeMux()
	authMiddleware := middleware.AuthMiddleware{
		Logger:       logger,
		SessionStore: sessionStore,
	}
	protected.HandleFunc(http.MethodGet+" /", h.Main)
	protected.HandleFunc(http.MethodGet+" /stats", h.GetStats)
	protected.HandleFunc(http.MethodPost+" /jobs", h.AddJob)
	protected.HandleFunc(http.MethodGet+" /jobs", h.GetJobs)
	protected.HandleFunc(http.MethodGet+" /jobs/{id}", h.JobDetails)
	protected.HandleFunc(http.MethodPatch+" /jobs/{id}", h.UpdateJob)
	protected.HandleFunc(http.MethodPost+" /jobs/{id}/notes", h.AddNote)
	protected.HandleFunc(http.MethodGet+" /signout", h.Signout)
	protected.HandleFunc(http.MethodGet+" /settings", h.Settings)
	protected.HandleFunc(http.MethodPost+" /settings/changePassword", h.ChangePassword)
	protected.HandleFunc(http.MethodPost+" /settings/logoutSessions", h.LogoutSessions)
	protected.HandleFunc(http.MethodPost+" /settings/deleteAccount", h.DeleteAccount)

	router.Handle("/", authMiddleware.Middleware(protected))

	loggingMiddleware := middleware.LoggingMiddleware{Logger: logger}
	return loggingMiddleware.Middleware(router)
}

func getVersion() string {
	version := os.Getenv("VERSION")
	if version == "" {
		version = "dev"
	}
	return version
}
