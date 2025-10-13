package router

import (
	"log/slog"
	"net/http"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/jobs/server/handler"
	"github.com/Piszmog/pathwise/internal/server/health"
	"github.com/Piszmog/pathwise/internal/server/middleware"
	"github.com/Piszmog/pathwise/internal/server/mux"
)

func New(logger *slog.Logger, database db.Database) http.Handler {
	h := handler.Handler{
		Logger:   logger,
		Database: database,
	}
	loggingMiddleware := middleware.LoggingMiddleware{Logger: logger}
	correlationMiddleware := middleware.CorrelationIDMiddleware{}

	return correlationMiddleware.Middleware(
		loggingMiddleware.Middleware(
			mux.NewMux(
				mux.WithHandleFunc(http.MethodPost, "/search", h.Search),
				mux.WithHandleFunc(http.MethodGet, "/health", health.Handle(logger)),
			),
		),
	)
}
