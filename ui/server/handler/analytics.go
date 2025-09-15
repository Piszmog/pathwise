package handler

import (
	"net/http"

	"github.com/Piszmog/pathwise/ui/components"
	"github.com/Piszmog/pathwise/ui/types"
)

func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	h.html(r.Context(), w, http.StatusOK, components.Analytics())
}

func (h *Handler) AnalyticsGraph(w http.ResponseWriter, r *http.Request) {
	sankeyData := getSampleSankeyData()
	h.html(r.Context(), w, http.StatusOK, components.SankeyGraph(sankeyData))
}

func getSampleSankeyData() types.SankeyData {
	return types.SankeyData{
		Nodes: []types.SankeyNode{
			{Name: "Applied"},
			{Name: "Watching"},
			{Name: "Interviewing"},
			{Name: "Offered"},
			{Name: "Accepted"},
			{Name: "Rejected"},
			{Name: "Withdrawn"},
			{Name: "Declined"},
		},
		Links: []types.SankeyLink{
			{Source: 0, Target: 2, Value: 25},
			{Source: 0, Target: 5, Value: 45},
			{Source: 0, Target: 6, Value: 8},
			{Source: 1, Target: 0, Value: 15},
			{Source: 1, Target: 6, Value: 3},
			{Source: 2, Target: 3, Value: 8},
			{Source: 2, Target: 5, Value: 17},
			{Source: 3, Target: 4, Value: 5},
			{Source: 3, Target: 7, Value: 3},
		},
	}
}
