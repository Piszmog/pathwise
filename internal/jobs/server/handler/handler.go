package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"

	contextkey "github.com/Piszmog/pathwise/internal/context_key"
	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/Piszmog/pathwise/internal/search"
)

type Handler struct {
	Logger   *slog.Logger
	Database db.Database
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req search.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(r.Context(), w, http.StatusBadRequest, "request body is not valid JSON", err)
		return
	}
	h.Logger.DebugContext(r.Context(), "recieved search request", "request", req)

	limit := int64(req.PerPage)
	offset := int64(req.Page) * limit
	keywords := req.Keywords

	if len(keywords) == 0 {
		keywords = []string{""}
	}

	results := make(map[string]search.JobListing)
	for _, keyword := range keywords {
		h.Logger.DebugContext(r.Context(), "searching for job listings", "keyword", keyword)
		param := queries.SearchHNJobsParams{
			Title:    db.NewNullString(req.Title),
			Company:  db.NewNullString(req.Company),
			Location: db.NewNullString(req.Location),
			IsRemote: req.IsRemote,
			IsHybrid: req.IsHybrid,
			Keyword:  keyword,
			Limit:    limit,
			Offset:   offset,
		}

		res, err := h.Database.Queries().SearchHNJobs(r.Context(), param)
		if err != nil {
			h.writeError(r.Context(), w, http.StatusInternalServerError, "failed to search for HN Jobs", err)
			return
		}

		for _, r := range res {
			if _, ok := results[r.ID]; !ok {
				results[r.ID] = search.JobListing{
					ID:       r.ID,
					Title:    r.Title,
					Company:  r.Company,
					Location: r.Location.String,
					IsRemote: r.IsRemote == 1,
					IsHybrid: r.IsHybrid == 1,
					Posted:   r.Posted.Time,
				}
			}
		}
	}
	h.Logger.DebugContext(r.Context(), "found search requests", "results", results)

	listings := make([]search.JobListing, 0, len(results))
	for _, v := range results {
		listings = append(listings, v)
	}
	sort.Slice(listings, func(i, j int) bool {
		return listings[i].Posted.After(listings[j].Posted)
	})

	res := search.Response{JobListings: listings}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to encode response", "error", err)
		return
	}
}

func (h *Handler) writeError(ctx context.Context, w http.ResponseWriter, status int, message string, parentErr error) {
	id := ctx.Value(contextkey.KeyCorrelationID).(contextkey.Key)
	h.Logger.ErrorContext(ctx, "failed to handle request", "error", parentErr, "message", message, "id", id)

	res := search.Error{Message: message}
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.Logger.ErrorContext(ctx, "failed to encode error response", "error", err, "status", status, "message", message, "id", id)
	}
}
