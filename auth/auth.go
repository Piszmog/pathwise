package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/Piszmog/pathwise/db/store"
)

type SessionJanitor struct {
	Logger *slog.Logger
	Store  *store.SessionStore
}

func (j *SessionJanitor) Run() {
	ctx := context.Background()
	j.Logger.Info("starting session janitor")
	for {
		j.Logger.Debug("deleting expired sessions")
		err := j.Store.DeleteExpired(ctx)
		if err != nil {
			j.Logger.Error("failed to delete expired sessions", "error", err)
		}
		j.Logger.Debug("deleted expired sessions")
		j.Logger.Debug("sleeping for 1 hour")
		time.Sleep(time.Hour)
	}
}
