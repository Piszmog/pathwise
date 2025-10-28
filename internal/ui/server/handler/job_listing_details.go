package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/db/queries"
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

	hasAdded := false
	userID, err := getUserID(r)
	if err == nil {
		_, err = h.Database.Queries().CheckUserHasAddedHNJob(r.Context(),
			queries.CheckUserHasAddedHNJobParams{UserID: userID, HnJobID: id})
		hasAdded = (err == nil)
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
		HasAdded:   hasAdded,
	}

	h.html(r.Context(), w, http.StatusOK, components.JobListingDetails(jobDetails))
}

func (h *Handler) AddJobApplicationFromListing(w http.ResponseWriter, r *http.Request) {
	jobListingID := r.PathValue("id")

	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get user id", "error", err)
		h.html(r.Context(), w, http.StatusUnauthorized,
			components.Alert(types.AlertTypeError, "Unauthorized", "Please sign in."))
		return
	}

	_, err = h.Database.Queries().CheckUserHasAddedHNJob(r.Context(),
		queries.CheckUserHasAddedHNJobParams{UserID: userID, HnJobID: jobListingID})
	if err == nil {
		h.html(r.Context(), w, http.StatusOK, components.JobListingAdded())
		return
	}

	hnJob, err := h.Database.Queries().GetHNJobByID(r.Context(), jobListingID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get hn job", "error", err)
		h.html(r.Context(), w, http.StatusNotFound,
			components.Alert(types.AlertTypeError, "Job not found", "This job listing no longer exists."))
		return
	}

	appURL := hnJob.ApplicationUrl.String
	if appURL == "" {
		appURL = hnJob.JobsUrl.String
	}

	tx, err := h.Database.DB().BeginTx(r.Context(), nil)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to begin transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	defer func() {
		if txErr := tx.Rollback(); txErr != nil {
			err = errors.Join(err, txErr)
		}
	}()

	qtx := queries.New(tx)

	jobApp := queries.InsertJobApplicationParams{
		Company:        hnJob.Company,
		Title:          hnJob.Title,
		Url:            db.NewNullString(appURL),
		UserID:         userID,
		SalaryMin:      sql.NullInt64{},
		SalaryMax:      sql.NullInt64{},
		SalaryCurrency: sql.NullString{},
	}

	jobID, err := qtx.InsertJobApplication(r.Context(), jobApp)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to insert job application", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	err = qtx.InsertJobApplicationStatusHistory(r.Context(), jobID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to insert status history", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	err = qtx.InsertUserHNJob(r.Context(), queries.InsertUserHNJobParams{
		UserID:           userID,
		HnJobID:          jobListingID,
		JobApplicationID: jobID,
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to insert user hn job", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	companyCount, err := h.Database.Queries().CountJobApplicationCompany(r.Context(),
		queries.CountJobApplicationCompanyParams{UserID: userID, Company: hnJob.Company})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to count company", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	companyIncrement := int64(0)
	if companyCount == 0 {
		companyIncrement = 1
	}

	err = qtx.IncrementNewJobApplicationStat(r.Context(), queries.IncrementNewJobApplicationStatParams{
		UserID:         userID,
		TotalCompanies: companyIncrement,
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to increment stats", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = tx.Commit(); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to commit transaction", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.JobListingAdded())
}
