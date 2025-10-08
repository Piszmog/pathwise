package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/Piszmog/pathwise/internal/ui/components"
	"github.com/Piszmog/pathwise/internal/ui/types"
)

func (h *Handler) GetJobListingDetails(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	hnJob, err := h.Database.Queries().GetHNJobByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.html(r.Context(), w, http.StatusNotFound,
				components.Alert(types.AlertTypeError, "Job not found", "The requested job could not be found."))
			return
		}
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Error", "Failed to load job details."))
		return
	}

	techStacks, err := h.Database.Queries().GetHNJobTechStacks(r.Context(), id)
	if err != nil {
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Error", "Failed to load job tech stacks."))
		return
	}
	appURL := hnJob.ApplicationUrl.String
	if appURL == "" {
		appURL = hnJob.JobsUrl.String
	}

	sourceID := strconv.FormatInt(hnJob.HnCommentID, 10)

	jobDetails := types.JobListingDetails{
		JobListing: types.JobListing{
			ID:                 hnJob.ID,
			Source:             types.JobSourceHackerNews,
			SourceID:           sourceID,
			SourceURL:          "https://news.ycombinator.com/item?id=" + sourceID,
			Company:            hnJob.Company,
			CompanyDescription: hnJob.CompanyDescription,
			Title:              hnJob.Title,
			CompanyURL:         hnJob.CompanyUrl.String,
			ContactEmail:       hnJob.ContactEmail.String,
			Description:        hnJob.Description.String,
			RoleType:           hnJob.RoleType.String,
			Location:           hnJob.Location.String,
			Salary:             hnJob.Salary.String,
			Equity:             hnJob.Equity.String,
			IsHybrid:           hnJob.IsHybrid != 0,
			IsRemote:           hnJob.IsRemote != 0,
			ApplicationURL:     appURL,
		},
		TechStacks: techStacks,
	}

	h.html(r.Context(), w, http.StatusOK, components.JobListingDetails(jobDetails))
}
