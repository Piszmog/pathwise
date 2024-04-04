package store

import (
	"context"
	"math"
	"strconv"

	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
)

type StatsStore struct {
	Database db.Database
}

func (s *StatsStore) Get(ctx context.Context, userID int) (types.StatsOpts, error) {
	row := s.Database.DB().QueryRowContext(ctx, statsGetQuery, userID, userID, userID)
	var totalApplications int
	var totalCompanies int
	var averageTimeToHearBack float64
	var totalInterviewing int
	var totalRejections int
	err := row.Scan(
		&totalApplications,
		&totalCompanies,
		&averageTimeToHearBack,
		&totalInterviewing,
		&totalRejections,
	)
	if err != nil {
		return types.StatsOpts{}, err
	}
	return types.StatsOpts{
		TotalApplications:           strconv.Itoa(totalApplications),
		TotalCompanies:              strconv.Itoa(totalCompanies),
		AverageTimeToHearBackInDays: strconv.FormatFloat(averageTimeToHearBack, 'f', 0, 64),
		TotalInterviewingPercentage: strconv.FormatFloat(math.Ceil((float64(totalInterviewing)/float64(totalApplications))*100), 'f', 0, 64),
		TotalRejectionsPercentage:   strconv.FormatFloat(math.Ceil((float64(totalRejections)/float64(totalApplications))*100), 'f', 0, 64),
	}, nil
}

const statsGetQuery = `
WITH first_status AS (
    SELECT
        job_application_id,
        MIN(created_at) AS first_status_at
    FROM job_application_status_histories
    WHERE status IN ('interviewing', 'rejected', 'cancelled', 'closed')
    GROUP BY job_application_id
)
SELECT
    COUNT(*) AS total_applications,
    COUNT(DISTINCT ja.company) AS total_companies,
    ROUND(IFNULL(AVG(JULIANDAY(fs.first_status_at) - JULIANDAY(ja.applied_at)), 0), 0) AS average_time_to_hear_back,
    COALESCE(SUM(CASE WHEN ja.status = 'interviewing' THEN 1 ELSE 0 END), 0) AS total_interviewing,
    COALESCE(SUM(CASE WHEN ja.status = 'rejected' THEN 1 ELSE 0 END), 0) AS total_rejections
FROM job_applications AS ja
LEFT JOIN first_status AS fs ON ja.id = fs.job_application_id
WHERE ja.user_id = ?
`
