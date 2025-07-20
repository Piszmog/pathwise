package handler

import (
	"fmt"
	"net/http"

	"github.com/Piszmog/pathwise/version"
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.Database.DB().PingContext(r.Context()); err != nil {
		h.Logger.Error("Health check failed", "error", err)
		http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}

	response := fmt.Sprintf(`{"status":"ok","database":"connected","version":"%s"}`, version.Value)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(response)); err != nil {
		h.Logger.Error("Failed to write health response", "error", err)
	}
}
