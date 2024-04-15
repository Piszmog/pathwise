package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Piszmog/pathwise/db"
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

	if err := t.Render(context.Background(), w); err != nil {
		h.Logger.Error("failed to render template", "error", err)
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
