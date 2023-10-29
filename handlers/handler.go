package handlers

import (
	"context"
	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/types"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type Handler struct {
	JobApplicationStore              *store.JobApplicationStore
	JobApplicationNoteStore          *store.JobApplicationNoteStore
	JobApplicationStatusHistoryStore *store.JobApplicationStatusHistoryStore
	StatsStore                       *store.StatsStore
}

func (h *Handler) Jobs(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	pageQuery := queries.Get("page")
	perPageQuery := queries.Get("per_page")
	page := 0
	var err error
	if pageQuery != "" {
		page, err = strconv.Atoi(pageQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	perPage := 10
	if perPageQuery != "" {
		perPage, err = strconv.Atoi(perPageQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	jobs, err := h.JobApplicationStore.Get(r.Context(), store.GetOpts{Page: page, PerPage: perPage})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	statsOpts, err := h.StatsStore.Get(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	m := components.Main(
		jobs,
		statsOpts,
	)
	m.Render(r.Context(), w)
}

func (h *Handler) JobDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	job, err := h.JobApplicationStore.GetByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	timelineEntries, err := h.getTimelineEntries(r.Context(), id)
	if err != nil {
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	if company == "" || title == "" || url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	job := types.JobApplication{
		Company: company,
		Title:   title,
		URL:     url,
	}
	if err := h.JobApplicationStore.Insert(r.Context(), job); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jobs, err := h.JobApplicationStore.Get(r.Context(), store.GetOpts{Page: 0, PerPage: 10})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stats, err := h.StatsStore.Get(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	components.AddJob(jobs, stats).Render(r.Context(), w)
}

func (h *Handler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	status := r.FormValue("status")
	if company == "" || title == "" || url == "" || status == "" {
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	job.UpdatedAt = updatedAt

	var stats *types.StatsOpts
	var newTimelineEntry types.NewTimelineEntry
	if previousStatus != status {
		updatedStats, err := h.StatsStore.Get(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stats = &updatedStats

		latestStatus, err := h.JobApplicationStatusHistoryStore.GetLatestByID(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if latestStatus.ID > 0 {
			newTimelineEntry = types.NewTimelineEntry{
				SwapOOB: "beforebegin:#" + newTimelineID(firstTimelineEntryType, firstTimelineEntryID),
				Entry:   latestStatus,
			}
		}
	}

	components.UpdateJob(job, stats, newTimelineEntry).Render(r.Context(), w)
}

func newTimelineID(entryType string, entryID string) string {
	return strings.ToLower(entryType) + "-" + entryID + "-entry"
}

func (h *Handler) AddNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	note := r.FormValue("note")
	if note == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	firstTimelineEntryID := r.FormValue("firstTimelineEntryID")
	firstTimelineEntryType := r.FormValue("firstTimelineEntryType")

	jobNote := types.JobApplicationNote{
		JobApplicationID: id,
		Note:             note,
	}
	n, err := h.JobApplicationNoteStore.Insert(r.Context(), jobNote)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	components.AddNote(
		types.NewTimelineEntry{
			SwapOOB: "beforebegin:#" + newTimelineID(firstTimelineEntryType, firstTimelineEntryID),
			Entry:   n,
		},
	).Render(r.Context(), w)
}
