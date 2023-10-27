package store

import (
	"context"
	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
	"math"
	"strconv"
)

type StatsStore struct {
	Database db.Database
}

func (s *StatsStore) Get(ctx context.Context) (types.StatsOpts, error) {
	row := s.Database.DB().QueryRowContext(
		ctx,
		`
		WITH first_status AS (
			SELECT job_application_id, min(created_at) as first_status_at
			FROM job_application_status_histories
			WHERE status IN ('interviewing', 'rejected', 'cancelled', 'closed')
			GROUP BY job_application_id
		)
		SELECT
			(SELECT COUNT(*) FROM job_applications) as total_applications,
			(SELECT COUNT(DISTINCT company) FROM job_applications) as total_companies,
			ROUND(AVG(JULIANDAY(fs.first_status_at) - JULIANDAY(ja.applied_at))) as average_time_to_hear_back,
			SUM(CASE WHEN ja.status = 'interviewing' THEN 1 ELSE 0 END) as total_interviewing,
			SUM(CASE WHEN ja.status = 'rejected' THEN 1 ELSE 0 END) as total_rejections
		FROM job_applications ja
				 LEFT JOIN first_status fs ON ja.id = fs.job_application_id;`,
	)
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
