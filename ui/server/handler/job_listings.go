package handler

import (
	"net/http"

	"github.com/Piszmog/pathwise/ui/components"
	"github.com/Piszmog/pathwise/ui/types"
)

func (h *Handler) GetJobListings(w http.ResponseWriter, r *http.Request) {
	jobs := types.GetMockJobListings()
	h.html(r.Context(), w, http.StatusOK, components.JobListings(jobs))
}
