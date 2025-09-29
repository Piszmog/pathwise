package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/Piszmog/pathwise/ui/components"
	"github.com/Piszmog/pathwise/ui/types"
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

	var companyURL *string
	if hnJob.CompanyUrl.Valid {
		companyURL = &hnJob.CompanyUrl.String
	}

	var contactEmail *string
	if hnJob.ContactEmail.Valid {
		contactEmail = &hnJob.ContactEmail.String
	}

	var description *string
	if hnJob.Description.Valid {
		description = &hnJob.Description.String
	}

	var roleType *string
	if hnJob.RoleType.Valid {
		roleType = &hnJob.RoleType.String
	}

	var location *string
	if hnJob.Location.Valid {
		location = &hnJob.Location.String
	}

	var salary *string
	if hnJob.Salary.Valid {
		salary = &hnJob.Salary.String
	}

	var equity *string
	if hnJob.Equity.Valid {
		equity = &hnJob.Equity.String
	}

	jobDetails := types.JobListingDetails{
		JobListing: types.JobListing{
			ID:                 hnJob.ID,
			Source:             types.JobSourceHackerNews,
			SourceID:           strconv.FormatInt(hnJob.HnCommentID, 10),
			SourceURL:          nil,
			Company:            hnJob.Company,
			CompanyDescription: hnJob.CompanyDescription,
			Title:              hnJob.Title,
			CompanyURL:         companyURL,
			ContactEmail:       contactEmail,
			Description:        description,
			RoleType:           roleType,
			Location:           location,
			Salary:             salary,
			Equity:             equity,
			IsHybrid:           hnJob.IsHybrid != 0,
			IsRemote:           hnJob.IsRemote != 0,
		},
		TechStacks: techStacks,
	}

	h.html(r.Context(), w, http.StatusOK, components.JobListingDetails(jobDetails))
}
