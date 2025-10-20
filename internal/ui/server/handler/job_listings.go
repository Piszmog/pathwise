package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Piszmog/pathwise/internal/search"
	"github.com/Piszmog/pathwise/internal/ui/components"
	"github.com/Piszmog/pathwise/internal/ui/types"
)

func (h *Handler) GetJobListingsPage(w http.ResponseWriter, r *http.Request) {
	h.html(r.Context(), w, http.StatusOK, components.JobListingsMain())
}

func (h *Handler) GetJobListings(w http.ResponseWriter, r *http.Request) {
	page, perPage, err := getPageOpts(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get page opts", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest,
			components.Alert(types.AlertTypeError, "Error", "Invalid pagination parameters."))
		return
	}

	req, filterOpts, err := newJobListingSearchRequest(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get filter opts", "error", err)
		h.html(
			r.Context(),
			w,
			http.StatusBadRequest,
			components.Alert(types.AlertTypeError, "Error", "Invalid filter parameters."),
		)
		return
	}
	req.Page = page
	req.PerPage = perPage

	jobs, err := h.SearchClient.Search(r.Context(), req)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get filtered HN jobs", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError,
			components.Alert(types.AlertTypeError, "Error", "Failed to load job listings."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.JobListingsContent(jobs, types.PaginationOpts{
		Page:    page,
		PerPage: perPage,
		Showing: len(jobs),
	}, filterOpts))
}

func newJobListingSearchRequest(r *http.Request) (search.Request, types.JobListingFilterOpts, error) {
	queries := r.URL.Query()
	req := search.Request{}
	filterOpts := types.JobListingFilterOpts{}

	req.Title = queries.Get("title")
	req.Company = queries.Get("company")

	if keywords := queries.Get("keywords"); keywords != "" {
		k := strings.SplitSeq(keywords, ",")
		for v := range k {
			req.Keywords = append(req.Keywords, strings.TrimSpace(v))
		}
	}

	if techStack := queries.Get("tech_stack"); techStack != "" {
		filterOpts.TechStack = &techStack
	}

	if isRemoteStr := queries.Get("is_remote"); isRemoteStr != "" {
		isRemote, err := strconv.ParseBool(isRemoteStr)
		if err != nil {
			return req, filterOpts, err
		}
		req.IsRemote = isRemote
		filterOpts.IsRemote = &isRemote
	}

	if isHybridStr := queries.Get("is_hybrid"); isHybridStr != "" {
		isHybrid, err := strconv.ParseBool(isHybridStr)
		if err != nil {
			return req, filterOpts, err
		}
		req.IsHybrid = isHybrid
		filterOpts.IsHybrid = &isHybrid
	}

	return req, filterOpts, nil
}
