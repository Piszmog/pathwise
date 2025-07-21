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

	data, err := json.Marshal(response)
	if err != nil {
		h.Logger.Error("failed to marshal health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		h.Logger.Error("failed to write health response", "error", err)
	}
}
