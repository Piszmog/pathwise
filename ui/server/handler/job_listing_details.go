package handler

import (
	"net/http"

	"github.com/Piszmog/pathwise/ui/components"
	"github.com/Piszmog/pathwise/ui/types"
)

func (h *Handler) GetJobListingDetails(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	jobDetails := types.GetMockJobListingDetails(id)
	if jobDetails == nil {
		h.html(r.Context(), w, http.StatusNotFound,
			components.Alert(types.AlertTypeError, "Job not found", "The requested job could not be found."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.JobListingDetails(*jobDetails))
}
