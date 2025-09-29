package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/Piszmog/pathwise/ui/components"
	"github.com/Piszmog/pathwise/ui/types"
	"github.com/Piszmog/pathwise/ui/utils"
)

var (
	ErrInvalidHeardBackAtType = errors.New("invalid type for HeardBackAt field")
)

func (h *Handler) JobDetails(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse id", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	job, err := h.Database.Queries().GetJobApplicationByID(r.Context(), id)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get job", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	timelineEntries, err := h.getTimelineEntries(r.Context(), id)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get timeline entries", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	archived := job.Archived == 1
	j := types.JobApplication{
		ID:             job.ID,
		Company:        job.Company,
		Title:          job.Title,
		URL:            job.Url,
		Status:         types.JobApplicationStatus(job.Status),
		AppliedAt:      job.AppliedAt,
		UpdatedAt:      job.UpdatedAt,
		UserID:         job.UserID,
		Archived:       archived,
		SalaryMin:      job.SalaryMin,
		SalaryMax:      job.SalaryMax,
		SalaryCurrency: job.SalaryCurrency,
	}

	h.html(r.Context(), w, http.StatusOK, components.JobDetails(j, timelineEntries))
}

func (h *Handler) getTimelineEntries(ctx context.Context, id int64) ([]types.JobApplicationTimelineEntry, error) {
	notes, err := h.Database.Queries().GetJobApplicationNotesByJobApplicationID(ctx, id)
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
		h.Logger.ErrorContext(r.Context(), "failed to parse form", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	if company == "" || title == "" || url == "" {
		h.Logger.WarnContext(r.Context(), "missing required form values", "company", company, "title", title, "url", url)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Missing company, title, or url", "Please enter a company, title, and url."))
		return
	}

	// Parse salary fields
	salaryMinStr := r.FormValue("salary_min")
	salaryMaxStr := r.FormValue("salary_max")
	salaryCurrencyStr := r.FormValue("salary_currency")

	var salaryMin sql.NullInt64
	if salaryMinStr != "" {
		val, err := strconv.ParseInt(salaryMinStr, 10, 64)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to parse salary_min", "error", err, "value", salaryMinStr)
			h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Invalid salary min value", "Please enter a valid number for minimum salary."))
			return
		}
		salaryMin = sql.NullInt64{Int64: val, Valid: true}
	}

	var salaryMax sql.NullInt64
	if salaryMaxStr != "" {
		val, err := strconv.ParseInt(salaryMaxStr, 10, 64)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to parse salary_max", "error", err, "value", salaryMaxStr)
			h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Invalid salary max value", "Please enter a valid number for maximum salary."))
			return
		}
		salaryMax = sql.NullInt64{Int64: val, Valid: true}
	}

	salaryCurrency := sql.NullString{}
	if salaryCurrencyStr != "" {
		salaryCurrency = sql.NullString{String: salaryCurrencyStr, Valid: true}
	}

	// Validate salary range
	if salaryMin.Valid && salaryMax.Valid && salaryMin.Int64 > salaryMax.Int64 {
		h.Logger.WarnContext(r.Context(), "validation error: minimum salary cannot be greater than maximum salary", "min_salary", salaryMin.Int64, "max_salary", salaryMax.Int64)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Invalid Salary Range", "Minimum salary cannot be greater than maximum salary."))
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	tx, err := h.Database.DB().BeginTx(r.Context(), nil)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to begin transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	defer func() {
		if txErr := tx.Rollback(); txErr != nil {
			err = errors.Join(err, txErr)
		}
	}()

	qtx := queries.New(tx)

	job := queries.InsertJobApplicationParams{
		Company:        company,
		Title:          title,
		Url:            url,
		UserID:         userID,
		SalaryMin:      salaryMin,
		SalaryMax:      salaryMax,
		SalaryCurrency: salaryCurrency,
	}
	var jobID int64
	if jobID, err = qtx.InsertJobApplication(r.Context(), job); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to insert job", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	if err = qtx.InsertJobApplicationStatusHistory(r.Context(), jobID); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to insert job status history", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	companyCount, err := h.Database.Queries().CountJobApplicationCompany(r.Context(), queries.CountJobApplicationCompanyParams{UserID: userID, Company: company})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to count company", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	companyIncrement := int64(0)
	if companyCount == 0 {
		companyIncrement = 1
	}
	if err = qtx.IncrementNewJobApplicationStat(r.Context(), queries.IncrementNewJobApplicationStatParams{UserID: userID, TotalCompanies: companyIncrement}); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to increment new job application", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = tx.Commit(); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to commit transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	jobs, err := h.getJobApplicationsByUserID(r.Context(), userID, int64(0), defaultPerPage, defaultPage*defaultPerPage)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get jobs", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.JobsReload(jobs, types.PaginationOpts{Page: defaultPage, PerPage: defaultPerPage, Showing: len(jobs)}, types.FilterOpts{}))
}

func (h *Handler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse id", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	job, err := h.Database.Queries().GetJobApplicationByIDAndUserID(r.Context(), queries.GetJobApplicationByIDAndUserIDParams{ID: id, UserID: userID})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get job", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if job.UserID != userID {
		h.Logger.WarnContext(r.Context(), "user does not own job", "userID", userID, "jobUserID", job.UserID)
		h.html(r.Context(), w, http.StatusForbidden, components.Alert(types.AlertTypeError, "You do not have permission to update this job", "Try again later."))
		return
	}

	if err = r.ParseForm(); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse form", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	status := r.FormValue("status")

	// Parse salary fields
	salaryMinStr := r.FormValue("salary_min")
	salaryMaxStr := r.FormValue("salary_max")
	salaryCurrencyStr := r.FormValue("salary_currency")

	var salaryMin sql.NullInt64
	if salaryMinStr != "" {
		val, err := strconv.ParseInt(salaryMinStr, 10, 64)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to parse salary_min", "error", err, "value", salaryMinStr)
			h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Invalid salary min value", "Please enter a valid number for minimum salary."))
			return
		}
		salaryMin = sql.NullInt64{Int64: val, Valid: true}
	}

	var salaryMax sql.NullInt64
	if salaryMaxStr != "" {
		val, err := strconv.ParseInt(salaryMaxStr, 10, 64)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to parse salary_max", "error", err, "value", salaryMaxStr)
			h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Invalid salary max value", "Please enter a valid number for maximum salary."))
			return
		}
		salaryMax = sql.NullInt64{Int64: val, Valid: true}
	}

	salaryCurrency := sql.NullString{}
	if salaryCurrencyStr != "" {
		salaryCurrency = sql.NullString{String: salaryCurrencyStr, Valid: true}
	}
	// Validate salary range
	if salaryMin.Valid && salaryMax.Valid && salaryMin.Int64 > salaryMax.Int64 {
		h.Logger.WarnContext(r.Context(), "validation error: minimum salary cannot be greater than maximum salary", "min_salary", salaryMin.Int64, "max_salary", salaryMax.Int64)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Invalid Salary Range", "Minimum salary cannot be greater than maximum salary."))
		return
	}

	if company == "" || title == "" || url == "" || status == "" {
		h.Logger.WarnContext(r.Context(), "missing required form values", "company", company, "title", title, "url", url, "status", status)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Missing company, title, url, or status", "Please enter a company, title, url, and status."))
		return
	}
	previousStatus := r.FormValue("previousStatus")
	firstTimelineEntryID := r.FormValue("firstTimelineEntryID")
	firstTimelineEntryType := r.FormValue("firstTimelineEntryType")

	tx, err := h.Database.DB().BeginTx(r.Context(), nil)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to begin transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	defer func() {
		if txErr := tx.Rollback(); txErr != nil {
			err = errors.Join(err, txErr)
		}
	}()

	qtx := queries.New(tx)

	err = qtx.UpdateJobApplication(r.Context(), queries.UpdateJobApplicationParams{
		ID:             job.ID,
		Company:        company,
		Title:          title,
		Url:            url,
		Status:         status,
		SalaryMin:      salaryMin,
		SalaryMax:      salaryMax,
		SalaryCurrency: salaryCurrency,
		UserID:         userID,
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to update job", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	var daysSince int
	if types.ToJobApplicationStatus(job.Status) != types.ToJobApplicationStatus(status) {
		histories, countErr := h.Database.Queries().CountJobApplicationStatusHistoriesByJobApplicationID(r.Context(), job.ID)
		if countErr != nil {
			h.Logger.ErrorContext(r.Context(), "failed to count job application status histories", "error", err)
			h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
			return
		}
		if histories == 1 {
			daysSince = int(time.Since(job.AppliedAt).Hours() / 24)
		}

		err = qtx.InsertJobApplicationStatusHistoryWithStatus(r.Context(), queries.InsertJobApplicationStatusHistoryWithStatusParams{JobApplicationID: job.ID, Status: status})
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to insert job status history", "error", err)
			h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
			return
		}
	}

	statParams := getStatsDiff(types.JobApplicationStatus(previousStatus), types.JobApplicationStatus(status))
	statParams.UserID = userID

	if job.Company != company {
		currentCompanyCount, currentCountErr := h.Database.Queries().CountJobApplicationCompany(r.Context(), queries.CountJobApplicationCompanyParams{UserID: userID, Company: job.Company})
		if currentCountErr != nil {
			h.Logger.ErrorContext(r.Context(), "failed to count current company", "error", err)
			h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
			return
		}
		companyCount, countErr := h.Database.Queries().CountJobApplicationCompany(r.Context(), queries.CountJobApplicationCompanyParams{UserID: userID, Company: company})
		if countErr != nil {
			h.Logger.ErrorContext(r.Context(), "failed to count company", "error", err)
			h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
			return
		}

		if currentCompanyCount > 1 && companyCount == 0 {
			statParams.TotalCompanies = 1
		} else if currentCompanyCount == 1 && companyCount > 0 {
			statParams.TotalCompanies = -1
		}
	}

	if daysSince > 0 {
		statParams.AverageTimeToHearBack = int64(daysSince)
		statParams.AverageTimeToHearBack_2 = int64(daysSince)
	}

	statChanged := hasChanged(statParams)
	if statChanged {
		err = h.recalculateStats(r.Context(), qtx, userID)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to recalculate stats", "error", err)
			h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
			return
		}
	}

	if err = tx.Commit(); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to commit transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	var stats types.StatsOpts
	var newTimelineEntry types.NewTimelineEntry
	if previousStatus != status {
		latestStatus, latestErr := h.Database.Queries().GetLatestJobApplicationStatusHistoryByID(r.Context(), id)
		if latestErr != nil {
			h.Logger.ErrorContext(r.Context(), "failed to get latest status", "error", err)
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
	if statChanged {
		stats, err = h.getStats(r.Context(), userID)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to get stats", "error", err)
			h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
			return
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

func (h *Handler) ArchiveJobs(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse form", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	dateVal := r.FormValue("date")
	if dateVal == "" {
		h.Logger.WarnContext(r.Context(), "missing required form values", "date", dateVal)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Missing date", "Please enter a date."))
		return
	}

	date, err := time.Parse("2006-01-02", dateVal)
	if err != nil {
		h.Logger.WarnContext(r.Context(), "invalid date format", "date", dateVal)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Invalid date", "Please enter a date."))
		return
	}
	date = date.UTC().Add(24*time.Hour - time.Nanosecond)

	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	tx, err := h.Database.DB().BeginTx(r.Context(), nil)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to begin transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	defer func() {
		if txErr := tx.Rollback(); txErr != nil {
			err = errors.Join(err, txErr)
		}
	}()

	qtx := queries.New(tx)

	err = qtx.ArchiveJobApplications(r.Context(), queries.ArchiveJobApplicationsParams{UserID: userID, AppliedAt: date})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to archive job applications", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	jobsCount, err := qtx.CountJobApplicationsForStats(r.Context(), userID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to count job applications for stats", "userID", userID, "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	var jobs []queries.GetJobApplicationsForStatsRow
	if jobsCount > 0 {
		jobs, err = qtx.GetJobApplicationsForStats(r.Context(), userID)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to get job applications for stats", "userID", userID, "error", err)
			h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
			return
		}
	}

	companyCount, err := qtx.CountJobApplicationCompanies(r.Context(), queries.CountJobApplicationCompaniesParams{UserID: userID, Archived: 0})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get count application companies", "userID", userID, "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	statArgs := queries.SetJobApplicationStatParams{TotalCompanies: companyCount, UserID: userID}
	for _, j := range jobs {
		statArgs.TotalApplications += 1
		if j.HeardBackAt != nil {
			heardBackAt, err := time.Parse("2006-01-02 15:04:05", j.HeardBackAt.(string))
			if err != nil {
				h.Logger.ErrorContext(r.Context(), "failed to convert string to date", "heardBackAt", j.HeardBackAt, "error", err)
				h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
				return
			}
			diff := heardBackAt.Sub(j.AppliedAt)
			daysSince := int64(diff.Hours() / 24)

			if statArgs.AverageTimeToHearBack == 0 {
				statArgs.AverageTimeToHearBack = daysSince
			} else {
				statArgs.AverageTimeToHearBack = (daysSince + statArgs.AverageTimeToHearBack) / 2
			}
		}
		switch types.JobApplicationStatus(j.Status) {
		case types.JobApplicationStatusAccepted:
			statArgs.TotalAccepted += 1
		case types.JobApplicationStatusApplied:
			statArgs.TotalApplied += 1
		case types.JobApplicationStatusCanceled:
			statArgs.TotalCanceled += 1
		case types.JobApplicationStatusDeclined:
			statArgs.TotalDeclined += 1
		case types.JobApplicationStatusInterviewing:
			statArgs.TotalInterviewing += 1
		case types.JobApplicationStatusOffered:
			statArgs.TotalOffers += 1
		case types.JobApplicationStatusRejected:
			statArgs.TotalRejected += 1
		case types.JobApplicationStatusWatching:
			statArgs.TotalWatching += 1
		case types.JobApplicationStatusWithdrawn:
			statArgs.TotalWatching += 1
		case types.JobApplicationStatusClosed:
		}
	}

	err = qtx.SetJobApplicationStat(r.Context(), statArgs)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to set stats", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = tx.Commit(); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to commit transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	jobsPage, err := h.getJobApplicationsByUserID(r.Context(), userID, int64(0), defaultPerPage, defaultPage*defaultPerPage)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get jobs", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.JobsReload(jobsPage, types.PaginationOpts{Page: defaultPage, PerPage: defaultPerPage, Showing: len(jobsPage)}, types.FilterOpts{}))
}

func (h *Handler) UnarchiveJob(w http.ResponseWriter, r *http.Request) {
	jobIDStr := r.PathValue("id")
	jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse job id", "jobID", jobIDStr, "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Invalid job ID", "Please try again."))
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	tx, err := h.Database.DB().BeginTx(r.Context(), nil)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to begin transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	defer func() {
		if txErr := tx.Rollback(); txErr != nil {
			err = errors.Join(err, txErr)
		}
	}()

	qtx := queries.New(tx)

	err = qtx.UnarchiveJobApplication(r.Context(), queries.UnarchiveJobApplicationParams{
		ID:     jobID,
		UserID: userID,
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to unarchive job application", "jobID", jobID, "userID", userID, "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	// Recalculate stats after unarchiving
	err = h.recalculateStats(r.Context(), qtx, userID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to recalculate stats", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = tx.Commit(); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to commit transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	// Get the updated archived jobs list (archived = 1)
	jobs, err := h.getJobApplicationsByUserID(r.Context(), userID, int64(1), defaultPerPage, defaultPage*defaultPerPage)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get archived jobs", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.JobsReload(jobs, types.PaginationOpts{Page: defaultPage, PerPage: defaultPerPage, Showing: len(jobs)}, types.FilterOpts{}))
}

func (h *Handler) ArchiveJob(w http.ResponseWriter, r *http.Request) {
	jobIDStr := r.PathValue("id")
	jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse job id", "jobID", jobIDStr, "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Invalid job ID", "Please try again."))
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	tx, err := h.Database.DB().BeginTx(r.Context(), nil)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to begin transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	defer func() {
		if txErr := tx.Rollback(); txErr != nil {
			err = errors.Join(err, txErr)
		}
	}()

	qtx := queries.New(tx)

	err = qtx.ArchiveJobApplication(r.Context(), queries.ArchiveJobApplicationParams{
		ID:     jobID,
		UserID: userID,
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to archive job application", "jobID", jobID, "userID", userID, "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	// Recalculate stats after archiving
	err = h.recalculateStats(r.Context(), qtx, userID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to recalculate stats", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = tx.Commit(); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to commit transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	// Get the updated active jobs list (archived = 0)
	jobs, err := h.getJobApplicationsByUserID(r.Context(), userID, int64(0), defaultPerPage, defaultPage*defaultPerPage)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get active jobs", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.JobsReload(jobs, types.PaginationOpts{Page: defaultPage, PerPage: defaultPerPage, Showing: len(jobs)}, types.FilterOpts{}))
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

func getStatsDiff(currentStatus types.JobApplicationStatus, newStatus types.JobApplicationStatus) queries.UpdateJobApplicationStatParams {
	params := queries.UpdateJobApplicationStatParams{}
	switch currentStatus {
	case types.JobApplicationStatusAccepted:
		params.TotalAccepted = -1
	case types.JobApplicationStatusApplied:
		params.TotalApplied = -1
	case types.JobApplicationStatusCanceled:
		params.TotalCanceled = -1
	case types.JobApplicationStatusDeclined:
		params.TotalDeclined = -1
	case types.JobApplicationStatusInterviewing:
		params.TotalInterviewing = -1
	case types.JobApplicationStatusOffered:
		params.TotalOffers = -1
	case types.JobApplicationStatusRejected:
		params.TotalRejected = -1
	case types.JobApplicationStatusWatching:
		params.TotalWatching = -1
	case types.JobApplicationStatusWithdrawn:
		params.TotalWidthdrawn = -1
	case types.JobApplicationStatusClosed:
	}

	switch newStatus {
	case types.JobApplicationStatusAccepted:
		params.TotalAccepted = 1
	case types.JobApplicationStatusApplied:
		params.TotalApplied = 1
	case types.JobApplicationStatusCanceled:
		params.TotalCanceled = 1
	case types.JobApplicationStatusDeclined:
		params.TotalDeclined = 1
	case types.JobApplicationStatusInterviewing:
		params.TotalInterviewing = 1
	case types.JobApplicationStatusOffered:
		params.TotalOffers = 1
	case types.JobApplicationStatusRejected:
		params.TotalRejected = 1
	case types.JobApplicationStatusWatching:
		params.TotalWatching = 1
	case types.JobApplicationStatusWithdrawn:
		params.TotalWidthdrawn = 1
	case types.JobApplicationStatusClosed:
	}
	return params
}

func hasChanged(diff queries.UpdateJobApplicationStatParams) bool {
	return diff.TotalAccepted != 0 || diff.TotalApplied != 0 || diff.TotalCanceled != 0 || diff.TotalDeclined != 0 || diff.TotalInterviewing != 0 || diff.TotalOffers != 0 || diff.TotalRejected != 0 || diff.TotalWatching != 0 || diff.TotalWidthdrawn != 0 || diff.TotalCompanies != 0
}

func (h *Handler) recalculateStats(ctx context.Context, qtx *queries.Queries, userID int64) error {
	jobsCount, err := qtx.CountJobApplicationsForStats(ctx, userID)
	if err != nil {
		return err
	}

	var jobs []queries.GetJobApplicationsForStatsRow
	if jobsCount > 0 {
		jobs, err = qtx.GetJobApplicationsForStats(ctx, userID)
		if err != nil {
			return err
		}
	}

	companyCount, err := qtx.CountJobApplicationCompanies(ctx, queries.CountJobApplicationCompaniesParams{UserID: userID, Archived: 0})
	if err != nil {
		return err
	}

	statArgs := queries.SetJobApplicationStatParams{TotalCompanies: companyCount, UserID: userID}
	var allDays []int64

	for _, j := range jobs {
		statArgs.TotalApplications += 1
		if j.HeardBackAt != nil {
			var heardBackAt time.Time
			var err error

			switch v := j.HeardBackAt.(type) {
			case time.Time:
				heardBackAt = v
			case string:
				heardBackAt, err = time.Parse("2006-01-02 15:04:05", v)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unexpected type for HeardBackAt %T: %w", v, ErrInvalidHeardBackAtType)
			}
			diff := heardBackAt.Sub(j.AppliedAt)
			daysSince := int64(diff.Hours() / 24)
			allDays = append(allDays, daysSince)
		}
		switch types.JobApplicationStatus(j.Status) {
		case types.JobApplicationStatusAccepted:
			statArgs.TotalAccepted += 1
		case types.JobApplicationStatusApplied:
			statArgs.TotalApplied += 1
		case types.JobApplicationStatusCanceled:
			statArgs.TotalCanceled += 1
		case types.JobApplicationStatusDeclined:
			statArgs.TotalDeclined += 1
		case types.JobApplicationStatusInterviewing:
			statArgs.TotalInterviewing += 1
		case types.JobApplicationStatusOffered:
			statArgs.TotalOffers += 1
		case types.JobApplicationStatusRejected:
			statArgs.TotalRejected += 1
		case types.JobApplicationStatusWatching:
			statArgs.TotalWatching += 1
		case types.JobApplicationStatusWithdrawn:
			statArgs.TotalWatching += 1
		case types.JobApplicationStatusClosed:
		}
	}

	if len(allDays) > 0 {
		sum := int64(0)
		for _, days := range allDays {
			sum += days
		}
		statArgs.AverageTimeToHearBack = sum / int64(len(allDays))
	}

	return qtx.SetJobApplicationStat(ctx, statArgs)
}
