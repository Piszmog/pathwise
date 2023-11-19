package store

import (
	"context"
	"database/sql"

	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
)

type JobApplicationStatusHistoryStore struct {
	Database db.Database
}

func (s *JobApplicationStatusHistoryStore) GetLatestByID(ctx context.Context, id int) (types.JobApplicationStatusHistory, error) {
	return scanJobApplicationStatusHistory(s.Database.DB().QueryRowContext(ctx, historyGetLatestByIDQuery, id))
}

const historyGetLatestByIDQuery = `
SELECT
	h.id, h.job_application_id, h.status, h.created_at
FROM 
	job_application_status_histories h
WHERE
	h.job_application_id = ?
ORDER BY h.created_at DESC
LIMIT 1
`

func scanJobApplicationStatusHistory(row *sql.Row) (types.JobApplicationStatusHistory, error) {
	var history types.JobApplicationStatusHistory
	var status string
	err := row.Scan(
		&history.ID,
		&history.JobApplicationID,
		&status,
		&history.CreatedAt,
	)
	history.Status = types.ToJobApplicationStatus(status)
	return history, err
}

func (s *JobApplicationStatusHistoryStore) GetAllByID(ctx context.Context, id int) ([]types.JobApplicationStatusHistory, error) {
	rows, err := s.Database.DB().QueryContext(ctx, historyGetAllByIDQuery, id)
	if err != nil {
		return nil, err
	}
	return scanJobApplicationStatusHistories(rows)
}

var historyGetAllByIDQuery = `
SELECT
	h.id, h.job_application_id, h.status, h.created_at
FROM 
	job_application_status_histories h
WHERE
	h.job_application_id = ?
ORDER BY h.created_at DESC
`

func scanJobApplicationStatusHistories(rows *sql.Rows) ([]types.JobApplicationStatusHistory, error) {
	defer rows.Close()
	var histories []types.JobApplicationStatusHistory
	for rows.Next() {
		var history types.JobApplicationStatusHistory
		var status string
		err := rows.Scan(
			&history.ID,
			&history.JobApplicationID,
			&status,
			&history.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		history.Status = types.ToJobApplicationStatus(status)
		histories = append(histories, history)
	}
	return histories, nil
}
