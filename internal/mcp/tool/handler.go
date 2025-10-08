package tool

import (
	"log/slog"

	"github.com/Piszmog/pathwise/internal/db"
)

type Handler struct {
	Logger   *slog.Logger
	Database db.Database
}
