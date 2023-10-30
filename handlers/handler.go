package handlers

import (
	"context"
	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/types"
	"github.com/Piszmog/pathwise/utils"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
)

type Handler struct {
	Logger                           *slog.Logger
	JobApplicationStore              *store.JobApplicationStore
	JobApplicationNoteStore          *store.JobApplicationNoteStore
	JobApplicationStatusHistoryStore *store.JobApplicationStatusHistoryStore
	StatsStore                       *store.StatsStore
}

func (h *Handler) Main(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.JobApplicationStore.Get(r.Context(), store.LimitOpts{Page: 0, PerPage: 10})
	if err != nil {
		h.Logger.Error("failed to get jobs", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	statsOpts, err := h.StatsStore.Get(r.Context())
	if err != nil {
		h.Logger.Error("failed to get stats", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	m := components.Main(
		jobs,
		statsOpts,
	)
	m.Render(r.Context(), w)
}

func (h *Handler) FilterJobs(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	pageQuery := queries.Get("page")
	perPageQuery := queries.Get("per_page")
	page := 0
	var err error
	if pageQuery != "" {
		page, err = strconv.Atoi(pageQuery)
		if err != nil {
			h.Logger.Error("failed to parse page query", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	perPage := 10
	if perPageQuery != "" {
		perPage, err = strconv.Atoi(perPageQuery)
		if err != nil {
			h.Logger.Error("failed to parse per_page query", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	company := queries.Get("company")
	status := queries.Get("status")

	jobs, err := h.JobApplicationStore.Filter(r.Context(), store.LimitOpts{Page: page, PerPage: perPage}, company, status)
	if err != nil {
		h.Logger.Error("failed to get jobs", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	components.Jobs(jobs).Render(r.Context(), w)
}

func (h *Handler) JobDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.Logger.Error("failed to parse id", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	job, err := h.JobApplicationStore.GetByID(r.Context(), id)
	if err != nil {
		h.Logger.Error("failed to get job", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	timelineEntries, err := h.getTimelineEntries(r.Context(), id)
	if err != nil {
		h.Logger.Error("failed to get timeline entries", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := components.JobDetails(job, timelineEntries)
	details.Render(r.Context(), w)
}

func (h *Handler) getTimelineEntries(ctx context.Context, id int) ([]types.JobApplicationTimelineEntry, error) {
	notes, err := h.JobApplicationNoteStore.GetAllByID(ctx, id)
	if err != nil {
		return nil, err
	}
	histories, err := h.JobApplicationStatusHistoryStore.GetAllByID(ctx, id)
	if err != nil {
		return nil, err
	}
	timelineEntries := make([]types.JobApplicationTimelineEntry, len(notes)+len(histories))
	for i, note := range notes {
		timelineEntries[i] = note
	}
	for i, history := range histories {
		timelineEntries[i+len(notes)] = history
	}
	sort.Slice(timelineEntries, func(i, j int) bool {
		return timelineEntries[i].Created().After(timelineEntries[j].Created())
	})
	return timelineEntries, nil
}

func (h *Handler) AddJob(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	if company == "" || title == "" || url == "" {
		h.Logger.Error("missing required form values", "company", company, "title", title, "url", url)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	job := types.JobApplication{
		Company: company,
		Title:   title,
		URL:     url,
	}
	if err := h.JobApplicationStore.Insert(r.Context(), job); err != nil {
		h.Logger.Error("failed to insert job", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jobs, err := h.JobApplicationStore.Get(r.Context(), store.LimitOpts{Page: 0, PerPage: 10})
	if err != nil {
		h.Logger.Error("failed to get jobs", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stats, err := h.StatsStore.Get(r.Context())
	if err != nil {
		h.Logger.Error("failed to get stats", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	components.AddJob(jobs, stats).Render(r.Context(), w)
}

func (h *Handler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.Logger.Error("failed to parse id", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	status := r.FormValue("status")
	if company == "" || title == "" || url == "" || status == "" {
		h.Logger.Error("missing required form values", "company", company, "title", title, "url", url, "status", status)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	previousStatus := r.FormValue("previousStatus")
	firstTimelineEntryID := r.FormValue("firstTimelineEntryID")
	firstTimelineEntryType := r.FormValue("firstTimelineEntryType")

	job := types.JobApplication{
		ID:      id,
		Company: company,
		Title:   title,
		URL:     url,
		Status:  types.ToJobApplicationStatus(status),
	}
	updatedAt, err := h.JobApplicationStore.Update(r.Context(), job)
	if err != nil {
		h.Logger.Error("failed to update job", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	job.UpdatedAt = updatedAt

	var stats *types.StatsOpts
	var newTimelineEntry types.NewTimelineEntry
	if previousStatus != status {
		updatedStats, err := h.StatsStore.Get(r.Context())
		if err != nil {
			h.Logger.Error("failed to get stats", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stats = &updatedStats

		latestStatus, err := h.JobApplicationStatusHistoryStore.GetLatestByID(r.Context(), id)
		if err != nil {
			h.Logger.Error("failed to get latest status", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if latestStatus.ID > 0 {
			newTimelineEntry = types.NewTimelineEntry{
				SwapOOB: "beforebegin:#" + newTimelineID(types.ToJobApplicationTimelineType(firstTimelineEntryType), firstTimelineEntryID),
				Entry:   latestStatus,
			}
		}
	}

	components.UpdateJob(job, stats, newTimelineEntry).Render(r.Context(), w)
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

func (h *Handler) AddNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.Logger.Error("failed to parse id", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	note := r.FormValue("note")
	if note == "" {
		h.Logger.Error("missing required form values", "note", note)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jobNote := types.JobApplicationNote{
		JobApplicationID: id,
		Note:             note,
	}
	n, err := h.JobApplicationNoteStore.Insert(r.Context(), jobNote)
	if err != nil {
		h.Logger.Error("failed to insert note", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	components.TimelineEntry(n, false).Render(r.Context(), w)
}
