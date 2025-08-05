package server

import (
	"context"
	"fmt"
	"log/slog"
)

type logger struct {
	l *slog.Logger
}

func (l *logger) Infof(format string, v ...any) {
	l.l.InfoContext(context.Background(), fmt.Sprintf(format, v...))
}
func (l *logger) Errorf(format string, v ...any) {
	l.l.ErrorContext(context.Background(), fmt.Sprintf(format, v...))
}
