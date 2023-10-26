package store

import (
	"context"
	"database/sql"
	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
	"strings"
)

type JobApplicationStatusHistoryStore struct {
	Database db.Database
}

func (s *JobApplicationStatusHistoryStore) GetByID(ctx context.Context, id int) (types.JobApplicationStatusHistory, error) {
	row := s.Database.DB().QueryRowContext(
		ctx,
		`
		SELECT
		    h.id, h.job_application_id, h.status, h.created_at
		FROM 
		    job_application_status_histories h
		WHERE
		    h.id = ?`,
		id,
	)
	return scanJobApplicationStatusHistory(row)
}

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

func (s *JobApplicationStatusHistoryStore) Get(ctx context.Context, opts GetOpts) ([]types.JobApplicationStatusHistory, error) {
	rows, err := s.Database.DB().QueryContext(
		ctx,
		`
		SELECT
		    h.id, h.job_application_id, h.status, h.created_at
		FROM 
		    job_application_status_histories h
		ORDER BY h.created_at DESC
		LIMIT ? OFFSET ?`,
		opts.PerPage,
		opts.Page*opts.PerPage,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobApplicationStatusHistories(rows)
}

func (s *JobApplicationStatusHistoryStore) GetAll(ctx context.Context) ([]types.JobApplicationStatusHistory, error) {
	rows, err := s.Database.DB().QueryContext(
		ctx,
		`
		SELECT
		    h.id, h.job_application_id, h.status, h.created_at
		FROM 
		    job_application_status_histories h
		ORDER BY h.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobApplicationStatusHistories(rows)
}

func scanJobApplicationStatusHistories(rows *sql.Rows) ([]types.JobApplicationStatusHistory, error) {
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

func (s *JobApplicationStatusHistoryStore) Insert(ctx context.Context, rec types.JobApplicationStatusHistory) error {
	_, err := s.Database.DB().ExecContext(
		ctx,
		`INSERT INTO job_application_status_histories (job_application_id, status) VALUES (?, ?)`,
		rec.JobApplicationID,
		strings.ToLower(rec.Status.String()),
	)
	return err
}

func (s *JobApplicationStatusHistoryStore) Update(ctx context.Context, rec types.JobApplicationStatusHistory) error {
	_, err := s.Database.DB().ExecContext(
		ctx,
		`UPDATE job_application_status_histories SET status = ? WHERE id = ?`,
		strings.ToLower(rec.Status.String()),
	)
	return err
}

func (s *JobApplicationStatusHistoryStore) Delete(ctx context.Context, id int) error {
	_, err := s.Database.DB().ExecContext(
		ctx,
		`DELETE FROM job_application_status_histories WHERE id = ?`,
		id,
	)
	return err
}
