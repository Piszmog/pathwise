package middleware

import (
	"log/slog"
	"net/http"
)

type LoggingMiddleware struct {
	Logger *slog.Logger
}

func (m *LoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.Logger.Debug("request received", "method", r.Method, "requestURI", r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
