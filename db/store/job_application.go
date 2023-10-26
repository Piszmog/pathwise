package store

import (
	"context"
	"database/sql"
	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
	"strings"
)

type JobApplicationStore struct {
	Database db.Database
}

func (s *JobApplicationStore) GetByID(ctx context.Context, id int) (types.JobApplication, error) {
	row := s.Database.DB().QueryRowContext(
		ctx,
		`
		SELECT
		    j.id, j.company, j.title, j.url, j.status, j.applied_at, j.updated_at
		FROM 
		    job_applications j
		WHERE
		    j.id = ?`,
		id,
	)
	return scanJobApplication(row)
}

func scanJobApplication(row *sql.Row) (types.JobApplication, error) {
	var job types.JobApplication
	var status string
	err := row.Scan(
		&job.ID,
		&job.Company,
		&job.Title,
		&job.URL,
		&status,
		&job.AppliedAt,
		&job.UpdatedAt,
	)
	job.Status = types.ToJobApplicationStatus(status)
	return job, err
}

func (s *JobApplicationStore) Get(ctx context.Context, opts GetOpts) ([]types.JobApplication, error) {
	rows, err := s.Database.DB().QueryContext(
		ctx,
		`
		SELECT
		    j.id, j.company, j.title, j.url, j.status, j.applied_at, j.updated_at
		FROM 
		    job_applications j
		ORDER BY j.applied_at DESC
		LIMIT ? OFFSET ?`,
		opts.PerPage,
		opts.Page*opts.PerPage,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobApplications(rows)
}

func (s *JobApplicationStore) GetAllByID(ctx context.Context, id int) ([]types.JobApplication, error) {
	return nil, nil
}

func scanJobApplications(rows *sql.Rows) ([]types.JobApplication, error) {
	var jobs []types.JobApplication
	for rows.Next() {
		var job types.JobApplication
		var status string
		err := rows.Scan(
			&job.ID,
			&job.Company,
			&job.Title,
			&job.URL,
			&status,
			&job.AppliedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		job.Status = types.ToJobApplicationStatus(status)
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s *JobApplicationStore) Insert(ctx context.Context, rec types.JobApplication) error {
	_, err := s.Database.DB().ExecContext(
		ctx,
		`INSERT INTO job_applications (company, title, url) VALUES (?, ?, ?)`,
		rec.Company,
		rec.Title,
		rec.URL,
	)
	return err
}

func (s *JobApplicationStore) Update(ctx context.Context, rec types.JobApplication) error {
	_, err := s.Database.DB().ExecContext(
		ctx,
		`UPDATE job_applications SET company = ?, title = ?, url = ?, status = ? WHERE id = ?`,
		rec.Company,
		rec.Title,
		rec.URL,
		strings.ToLower(rec.Status.String()),
		rec.ID,
	)
	return err
}

func (s *JobApplicationStore) Delete(ctx context.Context, id int) error {
	_, err := s.Database.DB().ExecContext(
		ctx,
		`DELETE FROM job_applications WHERE id = ?`,
		id,
	)
	return err
}
