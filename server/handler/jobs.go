package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/db/queries"
	"github.com/Piszmog/pathwise/types"
)

func (h *Handler) GetJobs(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Bad request."))
		return
	}

	urlQueries := r.URL.Query()
	archivedStr := urlQueries.Get("archived")
	archived := false
	if archivedStr != "" {
		archived, err = strconv.ParseBool(archivedStr)
		if err != nil {
			h.Logger.Warn("failed to parse archived query", "error", err)
			h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Bad request."))
			return
		}
	}

	page, perPage, err := getPageOpts(r)
	if err != nil {
		h.Logger.Error("failed to get page opts", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Bad request."))
		return
	}
	filterOpts := getFilterOpts(r)

	jobs, total, err := h.filterJobs(r.Context(), userID, archived, filterOpts, page, perPage)
	if err != nil {
		h.Logger.Error("failed to get jobs", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.Jobs(
		jobs,
		types.PaginationOpts{Page: page, PerPage: perPage, Total: total},
		filterOpts,
	))
}

func getFilterOpts(r *http.Request) types.FilterOpts {
	queries := r.URL.Query()
	return types.FilterOpts{
		Company: queries.Get("company"),
		Status:  types.ToJobApplicationStatus(queries.Get("status")),
	}
}

func getPageOpts(r *http.Request) (int64, int64, error) {
	queries := r.URL.Query()
	pageQuery := queries.Get("page")
	perPageQuery := queries.Get("per_page")
	page := defaultPage
	var err error
	if pageQuery != "" {
		page, err = strconv.ParseInt(pageQuery, 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}
	perPage := defaultPerPage
	if perPageQuery != "" {
		perPage, err = strconv.ParseInt(perPageQuery, 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}
	return page, perPage, nil
}

func (h *Handler) filterJobs(ctx context.Context, userID int64, archived bool, filterOpts types.FilterOpts, page int64, perPage int64) ([]types.JobApplication, int64, error) {
	h.Logger.Debug("filtering jobs", "filterOpts", filterOpts)
	offset := page * perPage
	archivedVal := int64(0)
	if archived {
		archivedVal = int64(1)
	}
	switch {
	case filterOpts.Company != "" && filterOpts.Status != "":
		return h.getJobApplictionsByUserIDAndCompanyAndStatus(ctx, userID, archivedVal, "%"+filterOpts.Company+"%", filterOpts.Status, perPage, offset)
	case filterOpts.Company != "" && filterOpts.Status == "":
		return h.getJobApplicationsByUserIDAndCompany(ctx, userID, archivedVal, "%"+filterOpts.Company+"%", perPage, offset)
	case filterOpts.Company == "" && filterOpts.Status != "":
		return h.getJobApplicationsByUserIDAndStatus(ctx, userID, archivedVal, filterOpts.Status, perPage, offset)
	default:
		apps, err := h.getJobApplicationsByUserID(ctx, userID, archivedVal, perPage, offset)
		if err != nil {
			return nil, 0, err
		}
		var totalApps int64
		if !archived {
			stats, err := h.getStats(ctx, userID)
			if err != nil {
				return nil, 0, err
			}
			totalApps = stats.TotalApplications
		} else {
			total, err := h.Database.Queries().CountJobApplicationsByUserID(ctx, queries.CountJobApplicationsByUserIDParams{UserID: userID, Archived: archivedVal})
			if err != nil {
				return nil, 0, err
			}
			totalApps = total
		}
		return apps, totalApps, nil
	}
}

func (h *Handler) getJobApplicationsByUserIDAndCompany(ctx context.Context, userID int64, archived int64, company string, perPage, offset int64) ([]types.JobApplication, int64, error) {
	h.Logger.Debug("getting job applications by user id and company", "userID", userID, "company", company)
	j, err := h.Database.Queries().GetJobApplicationsByUserIDAndCompany(
		ctx,
		queries.GetJobApplicationsByUserIDAndCompanyParams{
			UserID:   userID,
			Company:  company,
			Archived: archived,
			Limit:    perPage,
			Offset:   offset,
		},
	)
	if err != nil {
		return nil, 0, err
	}
	jobs := make([]types.JobApplication, len(j))
	for i, job := range j {
		jobs[i] = types.JobApplication{
			ID:        job.ID,
			Company:   job.Company,
			Title:     job.Title,
			URL:       job.Url,
			Status:    types.ToJobApplicationStatus(job.Status),
			AppliedAt: job.AppliedAt,
			UpdatedAt: job.UpdatedAt,
		}
	}

	total, err := h.Database.Queries().CountJobApplicationsByUserIDAndCompany(ctx, queries.CountJobApplicationsByUserIDAndCompanyParams{
		UserID:  userID,
		Company: company,
	})
	if err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (h *Handler) getJobApplictionsByUserIDAndCompanyAndStatus(ctx context.Context, userID int64, archived int64, company string, status types.JobApplicationStatus, perPage, offset int64) ([]types.JobApplication, int64, error) {
	h.Logger.Debug("getting job applications by user id, company, and status", "userID", userID, "company", company, "status", status)
	j, err := h.Database.Queries().GetJobApplicationsByUserIDAndCompanyAndStatus(
		ctx,
		queries.GetJobApplicationsByUserIDAndCompanyAndStatusParams{
			UserID:   userID,
			Company:  company,
			Status:   status.String(),
			Archived: archived,
			Limit:    perPage,
			Offset:   offset,
		},
	)
	if err != nil {
		return nil, 0, err
	}
	jobs := make([]types.JobApplication, len(j))
	for i, job := range j {
		jobs[i] = types.JobApplication{
			ID:        job.ID,
			Company:   job.Company,
			Title:     job.Title,
			URL:       job.Url,
			Status:    types.ToJobApplicationStatus(job.Status),
			AppliedAt: job.AppliedAt,
			UpdatedAt: job.UpdatedAt,
		}
	}

	total, err := h.Database.Queries().CountJobApplicationsByUserIDAndCompanyAndStatus(ctx, queries.CountJobApplicationsByUserIDAndCompanyAndStatusParams{
		UserID:  userID,
		Company: company,
		Status:  status.String(),
	})
	if err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (h *Handler) getJobApplicationsByUserIDAndStatus(ctx context.Context, userID int64, archived int64, status types.JobApplicationStatus, perPage, offset int64) ([]types.JobApplication, int64, error) {
	h.Logger.Debug("getting job applications by user id and status", "userID", userID, "status", status)
	j, err := h.Database.Queries().GetJobApplicationsByUserIDAndStatus(
		ctx,
		queries.GetJobApplicationsByUserIDAndStatusParams{
			UserID:   userID,
			Status:   status.String(),
			Archived: archived,
			Limit:    perPage,
			Offset:   offset,
		},
	)
	if err != nil {
		return nil, 0, err
	}
	jobs := make([]types.JobApplication, len(j))
	for i, job := range j {
		jobs[i] = types.JobApplication{
			ID:        job.ID,
			Company:   job.Company,
			Title:     job.Title,
			URL:       job.Url,
			Status:    types.ToJobApplicationStatus(job.Status),
			AppliedAt: job.AppliedAt,
			UpdatedAt: job.UpdatedAt,
		}
	}
	total, err := h.Database.Queries().CountJobApplicationsByUserIDAndStatus(ctx, queries.CountJobApplicationsByUserIDAndStatusParams{
		UserID: userID,
		Status: status.String(),
	})
	if err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

func (h *Handler) getJobApplicationsByUserID(ctx context.Context, userID int64, archived int64, perPage, offset int64) ([]types.JobApplication, error) {
	h.Logger.Debug("getting job applications by user id", "userID", userID)
	j, err := h.Database.Queries().GetJobApplicationsByUserID(
		ctx,
		queries.GetJobApplicationsByUserIDParams{
			UserID:   userID,
			Archived: archived,
			Limit:    perPage,
			Offset:   offset,
		},
	)
	if err != nil {
		return nil, err
	}
	jobs := make([]types.JobApplication, len(j))
	for i, job := range j {
		jobs[i] = types.JobApplication{
			ID:        job.ID,
			Company:   job.Company,
			Title:     job.Title,
			URL:       job.Url,
			Status:    types.ToJobApplicationStatus(job.Status),
			AppliedAt: job.AppliedAt,
			UpdatedAt: job.UpdatedAt,
		}
	}

	return jobs, nil
}
