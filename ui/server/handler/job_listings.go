package handler

import (
	"net/http"

	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/Piszmog/pathwise/ui/components"
	"github.com/Piszmog/pathwise/ui/types"
)

func (h *Handler) GetJobListings(w http.ResponseWriter, r *http.Request) {
	page, perPage, err := getPageOpts(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get page opts", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest,
			components.Alert(types.AlertTypeError, "Error", "Invalid pagination parameters."))
		return
	}

	offset := page * perPage
	hnJobs, err := h.Database.Queries().GetHNJobsPaginated(r.Context(), queries.GetHNJobsPaginatedParams{
		Limit:  perPage,
		Offset: offset,
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get paginated HN jobs", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Error", "Failed to load job listings."))
		return
	}

	jobs := make([]types.JobListing, len(hnJobs))
	for i, hnJob := range hnJobs {
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

		jobs[i] = types.JobListing{
			ID:                 hnJob.ID,
			Source:             types.JobSourceHackerNews,
			SourceID:           "",
			SourceURL:          nil,
			Company:            hnJob.Company,
			CompanyDescription: "",
			Title:              hnJob.Title,
			CompanyURL:         nil,
			ContactEmail:       nil,
			Description:        nil,
			RoleType:           roleType,
			Location:           location,
			Salary:             salary,
			Equity:             nil,
			IsHybrid:           hnJob.IsHybrid != 0,
			IsRemote:           hnJob.IsRemote != 0,
		}
	}

	h.html(r.Context(), w, http.StatusOK, components.JobListings(jobs, types.PaginationOpts{
		Page:    page,
		PerPage: perPage,
		Showing: len(jobs),
	}))
}
