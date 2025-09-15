package handler

import (
	"context"
	"database/sql"
	"net/http"
	"sort"

	"github.com/Piszmog/pathwise/ui/components"
	"github.com/Piszmog/pathwise/ui/types"
)

func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	h.html(r.Context(), w, http.StatusOK, components.Analytics())
}

func (h *Handler) AnalyticsGraph(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Bad request."))
		return
	}

	analyticsData, err := h.getAnalyticsData(r.Context(), userID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get analytics data", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Unable to load analytics", "There was a problem retrieving your data. Please try again later."))
	}

	h.html(r.Context(), w, http.StatusOK, components.SankeyGraph(analyticsData))
}

func (h *Handler) getAnalyticsData(ctx context.Context, userID int64) (types.AnalyticsData, error) {
	sankeyData, err := h.getSankeyData(ctx, userID)
	if err != nil {
		return types.AnalyticsData{}, err
	}

	return types.AnalyticsData{
		SankeyData: sankeyData,
	}, nil
}

func (h *Handler) getSankeyData(ctx context.Context, userID int64) (types.SankeyData, error) {
	dbTransitions, err := h.Database.Queries().GetStatusTransitionsForUser(ctx, userID)
	if err != nil && err != sql.ErrNoRows {
		return types.SankeyData{}, err
	}

	dbStatusCounts, err := h.Database.Queries().GetCurrentStatusCounts(ctx, userID)
	if err != nil && err != sql.ErrNoRows {
		return types.SankeyData{}, err
	}

	var transitions []types.StatusTransition
	for _, t := range dbTransitions {
		transitions = append(transitions, types.StatusTransition{
			FromStatus:      t.FromStatus,
			ToStatus:        t.ToStatus,
			TransitionCount: t.TransitionCount,
		})
	}

	var statusCounts []types.StatusCount
	for _, s := range dbStatusCounts {
		statusCounts = append(statusCounts, types.StatusCount{
			Status: s.Status,
			Count:  s.Count,
		})
	}

	return buildSankeyFromTransitions(transitions, statusCounts), nil
}

func buildSankeyFromTransitions(transitions []types.StatusTransition, statusCounts []types.StatusCount) types.SankeyData {
	statusSet := make(map[string]bool)
	for _, t := range transitions {
		statusSet[t.FromStatus] = true
		statusSet[t.ToStatus] = true
	}
	for _, s := range statusCounts {
		statusSet[s.Status] = true
	}

	var statuses []string
	for status := range statusSet {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)

	var nodes []types.SankeyNode
	statusToIndex := make(map[string]int)
	for i, status := range statuses {
		nodes = append(nodes, types.SankeyNode{Name: status})
		statusToIndex[status] = i
	}

	var links []types.SankeyLink
	for _, t := range transitions {
		fromIndex := statusToIndex[t.FromStatus]
		toIndex := statusToIndex[t.ToStatus]
		links = append(links, types.SankeyLink{
			Source: fromIndex,
			Target: toIndex,
			Value:  int(t.TransitionCount),
		})
	}

	if len(links) == 0 && len(statusCounts) > 0 {
		for _, s := range statusCounts {
			if index, exists := statusToIndex[s.Status]; exists {
				// Create self-referencing links to show status counts
				links = append(links, types.SankeyLink{
					Source: index,
					Target: index,
					Value:  int(s.Count),
				})
			}
		}
	}

	return types.SankeyData{
		Nodes: nodes,
		Links: links,
	}
}
