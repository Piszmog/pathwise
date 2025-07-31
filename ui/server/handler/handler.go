package handler

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/Piszmog/pathwise/ui/db"
	"github.com/a-h/templ"
)

const (
	defaultPage    = int64(0)
	defaultPerPage = int64(10)
)

type Handler struct {
	Logger   *slog.Logger
	Database db.Database
}

func (h *Handler) html(ctx context.Context, w http.ResponseWriter, status int, t templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if err := t.Render(ctx, w); err != nil {
		h.Logger.ErrorContext(ctx, "failed to render template", "error", err)
	}
}

func (h *Handler) htmlStatic(ctx context.Context, w http.ResponseWriter, status int, html []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(html); err != nil {
		h.Logger.ErrorContext(ctx, "Failed to write response", "error", err)
	}
}

func getUserID(r *http.Request) (int64, error) {
	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ip := strings.Split(xff, ",")[0]
		ip = strings.TrimSpace(ip)
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
