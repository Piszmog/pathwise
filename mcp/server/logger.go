package server

import (
	"fmt"
	"log/slog"
)

type logger struct {
	l *slog.Logger
}

func (l *logger) Infof(format string, v ...any) {
	l.l.Info(fmt.Sprintf(format, v...))
}
func (l *logger) Errorf(format string, v ...any) {
	l.l.Error(fmt.Sprintf(format, v...))
}
