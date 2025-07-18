package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Piszmog/pathwise/utils"
)

func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.Error("failed to get user ID", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jobApplications, err := h.Database.Queries().GetAllJobApplicationsByUserID(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get job applications for export", "error", err, "userID", userID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set headers for CSV download
	filename := fmt.Sprintf("job-applications-%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write CSV header
	header := []string{
		"Company",
		"Job Title",
		"Status",
		"Min Salary",
		"Max Salary",
		"Currency",
		"URL",
		"Applied Date",
		"Last Updated",
	}
	if err := writer.Write(header); err != nil {
		h.Logger.Error("failed to write CSV header", "error", err)
		return
	}

	// Write job application data
	for _, job := range jobApplications {
		var minSalary, maxSalary, currency string

		if job.SalaryMin.Valid {
			minSalary = strconv.FormatInt(job.SalaryMin.Int64, 10)
		}

		if job.SalaryMax.Valid {
			maxSalary = strconv.FormatInt(job.SalaryMax.Int64, 10)
		}

		if job.SalaryCurrency.Valid {
			currency = job.SalaryCurrency.String
		}

		record := []string{
			job.Company,
			utils.CleanJobTitle(job.Title),
			job.Status,
			minSalary,
			maxSalary,
			currency,
			job.Url,
			job.AppliedAt.Format("2006-01-02"),
			job.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if err := writer.Write(record); err != nil {
			h.Logger.Error("failed to write CSV record", "error", err)
			return
		}
	}

	h.Logger.Info("CSV export completed", "userID", userID, "recordCount", len(jobApplications))
}
