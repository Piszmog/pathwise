package handler

import (
	"context"
	"net/http"
	"sort"
	"strconv"

	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/db/queries"
	"github.com/Piszmog/pathwise/types"
	"github.com/Piszmog/pathwise/utils"
)

func (h *Handler) JobDetails(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.Logger.Error("failed to parse id", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	job, err := h.Database.Queries().GetJobApplicationByID(r.Context(), int64(id))
	if err != nil {
		h.Logger.Error("failed to get job", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	timelineEntries, err := h.getTimelineEntries(r.Context(), id)
	if err != nil {
		h.Logger.Error("failed to get timeline entries", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	j := types.JobApplication{
		ID:        job.ID,
		Company:   job.Company,
		Title:     job.Title,
		URL:       job.Url,
		Status:    types.JobApplicationStatus(job.Status),
		AppliedAt: job.AppliedAt,
		UpdatedAt: job.UpdatedAt,
		UserID:    job.UserID,
	}

	h.html(r.Context(), w, http.StatusOK, components.JobDetails(j, timelineEntries))
}

func (h *Handler) getTimelineEntries(ctx context.Context, id int64) ([]types.JobApplicationTimelineEntry, error) {
	notes, err := h.Database.Queries().GetJobApplicationNotesByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	histories, err := h.Database.Queries().GetJobApplicationStatusHistoriesByJobApplicationID(ctx, id)
	if err != nil {
		return nil, err
	}
	timelineEntries := make([]types.JobApplicationTimelineEntry, len(notes)+len(histories))
	for i, note := range notes {
		timelineEntries[i] = types.JobApplicationNote{
			ID:               note.ID,
			JobApplicationID: note.JobApplicationID,
			Note:             note.Note,
			CreatedAt:        note.CreatedAt,
		}
	}
	for i, history := range histories {
		timelineEntries[i+len(notes)] = types.JobApplicationStatusHistory{
			ID:               history.ID,
			JobApplicationID: history.JobApplicationID,
			Status:           types.JobApplicationStatus(history.Status),
			CreatedAt:        history.CreatedAt,
		}
	}
	sort.Slice(timelineEntries, func(i, j int) bool {
		return timelineEntries[i].Created().After(timelineEntries[j].Created())
	})
	return timelineEntries, nil
}

func (h *Handler) AddJob(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	if company == "" || title == "" || url == "" {
		h.Logger.Warn("missing required form values", "company", company, "title", title, "url", url)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Missing company, title, or url", "Please enter a company, title, and url."))
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	tx, err := h.Database.DB().BeginTx(r.Context(), nil)
	if err != nil {
		h.Logger.Error("failed to begin transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	defer tx.Rollback()

	qtx := queries.New(tx)

	job := queries.InsertJobApplicationParams{
		Company: company,
		Title:   title,
		Url:     url,
		UserID:  userID,
	}
	var jobID int64
	if jobID, err = qtx.InsertJobApplication(r.Context(), job); err != nil {
		h.Logger.Error("failed to insert job", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	if err = qtx.InsertJobApplicationStatusHistory(r.Context(), jobID); err != nil {
		h.Logger.Error("failed to insert job status history", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	if err = tx.Commit(); err != nil {
		h.Logger.Error("failed to commit transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	jobs, total, err := h.getJobApplicationsByUserID(r.Context(), userID, defaultPerPage, defaultPage*defaultPerPage)
	if err != nil {
		h.Logger.Error("failed to get jobs", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	stats, err := h.getStats(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get stats", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.AddJob(jobs, stats, types.PaginationOpts{Page: defaultPage, PerPage: defaultPerPage, Total: total}, types.FilterOpts{}))
}

func (h *Handler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.Logger.Error("failed to parse id", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	job, err := h.Database.Queries().GetJobApplicationByIDAndUserID(r.Context(), queries.GetJobApplicationByIDAndUserIDParams{ID: id, UserID: userID})
	if err != nil {
		h.Logger.Error("failed to get job", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if job.UserID != userID {
		h.Logger.Warn("user does not own job", "userID", userID, "jobUserID", job.UserID)
		h.html(r.Context(), w, http.StatusForbidden, components.Alert(types.AlertTypeError, "You do not have permission to update this job", "Try again later."))
		return
	}

	if err = r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	status := r.FormValue("status")
	if company == "" || title == "" || url == "" || status == "" {
		h.Logger.Warn("missing required form values", "company", company, "title", title, "url", url, "status", status)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Missing company, title, url, or status", "Please enter a company, title, url, and status."))
		return
	}
	previousStatus := r.FormValue("previousStatus")
	firstTimelineEntryID := r.FormValue("firstTimelineEntryID")
	firstTimelineEntryType := r.FormValue("firstTimelineEntryType")

	tx, err := h.Database.DB().BeginTx(r.Context(), nil)
	if err != nil {
		h.Logger.Error("failed to begin transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	defer tx.Rollback()

	qtx := queries.New(tx)

	err = qtx.UpdateJobApplication(r.Context(), queries.UpdateJobApplicationParams{
		ID:      job.ID,
		Company: company,
		Title:   title,
		Url:     url,
		Status:  status,
	})
	if err != nil {
		h.Logger.Error("failed to update job", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	err = qtx.InsertJobApplicationStatusHistoryWithStatus(r.Context(), queries.InsertJobApplicationStatusHistoryWithStatusParams{JobApplicationID: job.ID, Status: status})
	if err != nil {
		h.Logger.Error("failed to insert job status history", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = tx.Commit(); err != nil {
		h.Logger.Error("failed to commit transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	var stats types.StatsOpts
	var newTimelineEntry types.NewTimelineEntry
	if previousStatus != status {
		stats, err = h.getStats(r.Context(), userID)
		if err != nil {
			h.Logger.Error("failed to get stats", "error", err)
			h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
			return
		}

		latestStatus, err := h.Database.Queries().GetLatestJobApplicationStatusHistoryByID(r.Context(), int64(id))
		if err != nil {
			h.Logger.Error("failed to get latest status", "error", err)
			h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
			return
		}
		if latestStatus.ID > 0 {
			newTimelineEntry = types.NewTimelineEntry{
				SwapOOB: "beforebegin:#" + newTimelineID(types.ToJobApplicationTimelineType(firstTimelineEntryType), firstTimelineEntryID),
				Entry: types.JobApplicationStatusHistory{
					ID:               latestStatus.ID,
					JobApplicationID: latestStatus.JobApplicationID,
					Status:           types.JobApplicationStatus(latestStatus.Status),
					CreatedAt:        latestStatus.CreatedAt,
				},
			}
		}
	}

	actualJob := types.JobApplication{
		ID:        job.ID,
		Company:   company,
		Title:     title,
		URL:       url,
		Status:    types.JobApplicationStatus(status),
		AppliedAt: job.AppliedAt,
		UpdatedAt: job.UpdatedAt,
		UserID:    job.UserID,
	}

	h.html(r.Context(), w, http.StatusOK, components.UpdateJob(actualJob, stats, newTimelineEntry))
}

func newTimelineID(entryType types.JobApplicationTimelineType, entryID string) string {
	switch entryType {
	case types.JobApplicationTimelineTypeStatus:
		return utils.TimelineStatusRowStringID(entryID)
	case types.JobApplicationTimelineTypeNote:
		return utils.TimelineNoteRowStringID(entryID)
	default:
		return "unknown"
	}
}
