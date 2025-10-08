package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/Piszmog/pathwise/internal/ui/components"
	"github.com/Piszmog/pathwise/internal/ui/types"
)

func (h *Handler) GetJobListings(w http.ResponseWriter, r *http.Request) {
	page, perPage, err := getPageOpts(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get page opts", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest,
			components.Alert(types.AlertTypeError, "Error", "Invalid pagination parameters."))
		return
	}

	filterOpts, err := getJobListingFilterOpts(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get filter opts", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest,
			components.Alert(types.AlertTypeError, "Error", "Invalid filter parameters."))
		return
	}

	jobs, err := h.filterHNJobs(r.Context(), filterOpts, page, perPage)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get filtered HN jobs", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Error", "Failed to load job listings."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.JobListings(jobs, types.PaginationOpts{
		Page:    page,
		PerPage: perPage,
		Showing: len(jobs),
	}, filterOpts))
}

func getJobListingFilterOpts(r *http.Request) (types.JobListingFilterOpts, error) {
	queries := r.URL.Query()
	opts := types.JobListingFilterOpts{}

	if isRemoteStr := queries.Get("is_remote"); isRemoteStr != "" {
		isRemote, err := strconv.ParseBool(isRemoteStr)
		if err != nil {
			return opts, err
		}
		opts.IsRemote = &isRemote
	}

	if isHybridStr := queries.Get("is_hybrid"); isHybridStr != "" {
		isHybrid, err := strconv.ParseBool(isHybridStr)
		if err != nil {
			return opts, err
		}
		opts.IsHybrid = &isHybrid
	}

	if techStack := queries.Get("tech_stack"); techStack != "" {
		opts.TechStack = &techStack
	}

	return opts, nil
}

func (h *Handler) filterHNJobs(ctx context.Context, filterOpts types.JobListingFilterOpts, page int64, perPage int64) ([]types.JobListing, error) {
	offset := page * perPage

	isRemoteCheck := int64(-1)
	isRemoteParam := int64(-1)
	if filterOpts.IsRemote != nil {
		if *filterOpts.IsRemote {
			isRemoteCheck = int64(1)
			isRemoteParam = int64(1)
		} else {
			isRemoteCheck = int64(0)
			isRemoteParam = int64(0)
		}
	}

	isHybridCheck := int64(-1)
	isHybridParam := int64(-1)
	if filterOpts.IsHybrid != nil {
		if *filterOpts.IsHybrid {
			isHybridCheck = int64(1)
			isHybridParam = int64(1)
		} else {
			isHybridCheck = int64(0)
			isHybridParam = int64(0)
		}
	}

	techStackCheck := ""
	techStackParam := ""
	if filterOpts.TechStack != nil {
		techStackCheck = *filterOpts.TechStack
		techStackParam = *filterOpts.TechStack
	}

	hnJobs, err := h.Database.Queries().GetHNJobsFiltered(ctx, queries.GetHNJobsFilteredParams{
		Column1:  isRemoteCheck,
		IsRemote: isRemoteParam,
		Column3:  isHybridCheck,
		IsHybrid: isHybridParam,
		Column5:  techStackCheck,
		Value:    techStackParam,
		Limit:    perPage,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	jobs := make([]types.JobListing, len(hnJobs))
	for i, hnJob := range hnJobs {
		jobs[i] = types.JobListing{
			ID:                 hnJob.ID,
			Source:             types.JobSourceHackerNews,
			SourceID:           "",
			SourceURL:          "",
			Company:            hnJob.Company,
			CompanyDescription: "",
			Title:              hnJob.Title,
			CompanyURL:         "",
			ContactEmail:       "",
			Description:        "",
			RoleType:           hnJob.RoleType.String,
			Location:           hnJob.Location.String,
			Salary:             hnJob.Salary.String,
			Equity:             "",
			IsHybrid:           hnJob.IsHybrid != 0,
			IsRemote:           hnJob.IsRemote != 0,
			PostedAt:           hnJob.CommentedAt.Time.Format("Jan, 2, 2006"),
		}
	}

	return jobs, nil
}
