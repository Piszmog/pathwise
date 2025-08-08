package router

import (
	"log/slog"
	"net/http"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/ui/dist"
	"github.com/Piszmog/pathwise/ui/server/handler"
	"github.com/Piszmog/pathwise/ui/server/middleware"
)

func New(logger *slog.Logger, database db.Database) http.Handler {
	h := &handler.Handler{
		Logger:   logger,
		Database: database,
	}

	router := http.NewServeMux()
	router.Handle(http.MethodGet+" /assets/", middleware.Cache(http.FileServer(http.FS(dist.AssetsDir))))
	router.HandleFunc(http.MethodGet+" /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/assets/img/favicon.ico", http.StatusSeeOther)
	})
	router.HandleFunc(http.MethodGet+" /health", h.Health)
	router.HandleFunc(http.MethodGet+" /signup", h.Signup)
	router.HandleFunc(http.MethodPost+" /signup", h.Register)
	router.HandleFunc(http.MethodGet+" /signin", h.Signin)
	router.HandleFunc(http.MethodPost+" /signin", h.Authenticate)

	protected := http.NewServeMux()
	authMiddleware := middleware.AuthMiddleware{
		Logger:   logger,
		Database: database,
	}
	protected.HandleFunc(http.MethodGet+" /", h.Main)
	protected.HandleFunc(http.MethodGet+" /stats", h.GetStats)
	protected.HandleFunc(http.MethodPost+" /jobs", h.AddJob)
	protected.HandleFunc(http.MethodGet+" /archives", h.Archives)
	protected.HandleFunc(http.MethodPatch+" /jobs/archive", h.ArchiveJobs)
	protected.HandleFunc(http.MethodGet+" /jobs", h.GetJobs)
	protected.HandleFunc(http.MethodGet+" /jobs/{id}", h.JobDetails)
	protected.HandleFunc(http.MethodPatch+" /jobs/{id}", h.UpdateJob)
	protected.HandleFunc(http.MethodPatch+" /jobs/{id}/archive", h.ArchiveJob)
	protected.HandleFunc(http.MethodPatch+" /jobs/{id}/unarchive", h.UnarchiveJob)
	protected.HandleFunc(http.MethodPost+" /jobs/{id}/notes", h.AddNote)
	protected.HandleFunc(http.MethodGet+" /signout", h.Signout)
	protected.HandleFunc(http.MethodGet+" /settings", h.Settings)
	protected.HandleFunc(http.MethodPost+" /settings/changePassword", h.ChangePassword)
	protected.HandleFunc(http.MethodPost+" /settings/logoutSessions", h.LogoutSessions)
	protected.HandleFunc(http.MethodPost+" /settings/deleteAccount", h.DeleteAccount)
	protected.HandleFunc(http.MethodPost+" /settings/mcp/auth", h.CreateMcpAuth)
	protected.HandleFunc(http.MethodPatch+" /settings/mcp/auth", h.RegenerateMcpAuth)
	protected.HandleFunc(http.MethodDelete+" /settings/mcp/auth", h.DeleteMcpAuth)
	protected.HandleFunc(http.MethodGet+" /export/csv", h.ExportCSV)

	router.Handle("/", authMiddleware.Middleware(protected))

	loggingMiddleware := middleware.LoggingMiddleware{Logger: logger}
	return loggingMiddleware.Middleware(router)
}
