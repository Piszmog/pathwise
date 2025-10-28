package health

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Piszmog/pathwise/internal/version"
)

func Handle(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := healthResponse{
			Status:  "ok",
			Version: version.Value,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode health response", "error", err)
			return
		}
	}
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
