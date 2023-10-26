package handlers

import (
	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/types"
	"net/http"
	"time"
)

func GetJob(w http.ResponseWriter, r *http.Request) {
	details := components.JobDetails(
		types.JobApplication{
			ID:        1,
			Company:   "Company 1",
			Title:     "Title 1",
			Status:    types.JobApplicationStatusClosed,
			AppliedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		[]types.JobApplicationTimelineEntry{
			types.JobApplicationStatusHistory{
				ID:               1,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				JobApplicationID: 1,
				Status:           types.JobApplicationStatusApplied,
			},
			types.JobApplicationNote{
				ID:               2,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				JobApplicationID: 1,
				Note:             "This is a note",
			},
		},
	)
	details.Render(r.Context(), w)
}
