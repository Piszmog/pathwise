package handler

import (
	"context"
	"math"
	"net/http"
	"strconv"

	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/types"
)

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Bad request."))
		return
	}

	stats, err := h.getStats(r.Context(), userID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get stats", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.Stats(stats, false, ""))
}

func (h *Handler) getStats(ctx context.Context, userID int64) (types.StatsOpts, error) {
	stats, err := h.Database.Queries().GetJobApplicationStat(ctx, userID)
	if err != nil {
		return types.StatsOpts{}, err
	}

	statsOpts := types.StatsOpts{
		TotalApplications:           stats.TotalApplications,
		TotalCompanies:              stats.TotalCompanies,
		AverageTimeToHearBackInDays: stats.AverageTimeToHearBack,
		TotalInterviewingPercentage: toPercentage(stats.TotalInterviewing, stats.TotalApplications),
		TotalRejectionsPercentage:   toPercentage(stats.TotalRejected, stats.TotalApplications),
	}
	return statsOpts, nil
}

func toPercentage(value, total int64) string {
	return strconv.FormatFloat(math.Ceil((float64(value)/float64(total))*100), 'f', 0, 64)
}
