package handlers

import (
	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/types"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
	"strconv"
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
	notes, err := h.JobApplicationNoteStore.GetAll(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	histories, err := h.JobApplicationStatusHistoryStore.GetAll(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
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

	details := components.JobDetails(job, timelineEntries)
	details.Render(r.Context(), w)
}
