package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

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

	results := make(map[string]search.JobListing)
	for _, keyword := range req.Keywords {
		param := queries.SearchHNJobsParams{
			Title:    db.NewNullString(req.Title),
			Company:  db.NewNullString(req.Company),
			Location: db.NewNullString(req.Location),
			IsRemote: req.IsRemote,
			IsHybrid: req.IsHybrid,
			Keyword:  keyword,
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
					Location: r.Location.String,
					IsRemote: r.IsRemote == 1,
					IsHybrid: r.IsHybrid == 1,
					Posted:   r.Posted,
				}
			}
		}
	}

	listings := make([]search.JobListing, 0, len(results))
	for _, v := range results {
		listings = append(listings, v)
	}

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
