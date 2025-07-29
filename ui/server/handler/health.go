package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Piszmog/pathwise/version"
)

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response := healthResponse{
		Status:  "ok",
		Version: version.Value,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to encode health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
