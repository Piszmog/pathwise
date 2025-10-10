package router

import (
	"log/slog"
	"net/http"

	"github.com/Piszmog/pathwise/internal/db"
	mw "github.com/Piszmog/pathwise/internal/server/middleware"
	"github.com/Piszmog/pathwise/internal/server/mux"
	"github.com/Piszmog/pathwise/internal/ui/dist"
	"github.com/Piszmog/pathwise/internal/ui/server/handler"
	"github.com/Piszmog/pathwise/internal/ui/server/middleware"
)

func New(logger *slog.Logger, database db.Database) http.Handler {
	h := &handler.Handler{
		Logger:   logger,
		Database: database,
	}
	authMiddleware := middleware.AuthMiddleware{
		Logger:   logger,
		Database: database,
	}
	loggingMiddleware := mw.LoggingMiddleware{Logger: logger}

	return loggingMiddleware.Middleware(
		mux.NewMux(
			mux.WithHandle(http.MethodGet, "/assets/", middleware.Cache(http.FileServer(http.FS(dist.AssetsDir)))),
			mux.WithHandleFunc(http.MethodGet, "/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/assets/img/favicon.ico", http.StatusSeeOther)
			}),
			mux.WithHandleFunc(http.MethodGet, "/health", h.Health),
			mux.WithHandleFunc(http.MethodGet, "/signup", h.Signup),
			mux.WithHandleFunc(http.MethodPost, "/signup", h.Register),
			mux.WithHandleFunc(http.MethodGet, "/signin", h.Signin),
			mux.WithHandleFunc(http.MethodPost, "/signin", h.Authenticate),
			mux.WithGeneralHandle(
				"/",
				authMiddleware.Middleware(
					mux.NewMux(
						mux.WithHandleFunc(http.MethodGet, "/", h.Main),
						mux.WithHandleFunc(http.MethodGet, "/job-listings", h.GetJobListings),
						mux.WithHandleFunc(http.MethodGet, "/job-listings/{id}", h.GetJobListingDetails),
						mux.WithHandleFunc(http.MethodGet, "/stats", h.GetStats),
						mux.WithHandleFunc(http.MethodPost, "/jobs", h.AddJob),
						mux.WithHandleFunc(http.MethodGet, "/archives", h.Archives),
						mux.WithHandleFunc(http.MethodPatch, "/jobs/archive", h.ArchiveJobs),
						mux.WithHandleFunc(http.MethodGet, "/jobs", h.GetJobs),
						mux.WithHandleFunc(http.MethodGet, "/jobs/{id}", h.JobDetails),
						mux.WithHandleFunc(http.MethodPatch, "/jobs/{id}", h.UpdateJob),
						mux.WithHandleFunc(http.MethodPatch, "/jobs/{id}/archive", h.ArchiveJob),
						mux.WithHandleFunc(http.MethodPatch, "/jobs/{id}/unarchive", h.UnarchiveJob),
						mux.WithHandleFunc(http.MethodPost, "/jobs/{id}/notes", h.AddNote),
						mux.WithHandleFunc(http.MethodGet, "/signout", h.Signout),
						mux.WithHandleFunc(http.MethodGet, "/settings", h.Settings),
						mux.WithHandleFunc(http.MethodPost, "/settings/changePassword", h.ChangePassword),
						mux.WithHandleFunc(http.MethodPost, "/settings/logoutSessions", h.LogoutSessions),
						mux.WithHandleFunc(http.MethodPost, "/settings/deleteAccount", h.DeleteAccount),
						mux.WithHandleFunc(http.MethodPost, "/settings/mcp/auth", h.CreateMcpAuth),
						mux.WithHandleFunc(http.MethodPatch, "/settings/mcp/auth", h.RegenerateMcpAuth),
						mux.WithHandleFunc(http.MethodDelete, "/settings/mcp/auth", h.DeleteMcpAuth),
						mux.WithHandleFunc(http.MethodGet, "/export/csv", h.ExportCSV),
						mux.WithHandleFunc(http.MethodGet, "/analytics", h.Analytics),
						mux.WithHandleFunc(http.MethodGet, "/analytics/graph", h.AnalyticsGraph),
					),
				),
			),
		),
	)
}
