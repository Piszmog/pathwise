package store

import (
	"context"
	"database/sql"
	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
	"strings"
	"time"
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

func (s *JobApplicationStore) Get(ctx context.Context, opts LimitOpts) ([]types.JobApplication, error) {
	rows, err := s.Database.DB().QueryContext(
		ctx,
		`
		SELECT
		    j.id, j.company, j.title, j.url, j.status, j.applied_at, j.updated_at
		FROM 
		    job_applications j
		ORDER BY j.updated_at DESC
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

func (s *JobApplicationStore) Filter(ctx context.Context, opts LimitOpts, company string, status string) ([]types.JobApplication, error) {
	query := `SELECT
		    j.id, j.company, j.title, j.url, j.status, j.applied_at, j.updated_at
		FROM 
		    job_applications j`
	if company != "" || status != "" {
		query += ` WHERE`
		if company != "" {
			query += ` j.company LIKE ?`
		}
		if status != "" {
			if company != "" {
				query += ` AND`
			}
			query += ` j.status LIKE ?`
		}
	}
	query += ` ORDER BY j.updated_at DESC LIMIT ? OFFSET ?`
	var queryArgs = []interface{}{}
	if company != "" {
		queryArgs = append(queryArgs, "%"+company+"%")
	}
	if status != "" {
		queryArgs = append(queryArgs, "%"+status+"%")
	}
	queryArgs = append(queryArgs, opts.PerPage, opts.Page*opts.PerPage)

	rows, err := s.Database.DB().QueryContext(
		ctx,
		query,
		queryArgs...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobApplications(rows)
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
	tx, err := s.Database.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(
		ctx,
		`INSERT INTO job_applications (company, title, url) VALUES (?, ?, ?)`,
		rec.Company,
		rec.Title,
		rec.URL,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO job_application_status_histories (job_application_id) VALUES (?)`,
		id,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *JobApplicationStore) Update(ctx context.Context, rec types.JobApplication) (time.Time, error) {
	tx, err := s.Database.DB().BeginTx(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}
	_, err = tx.ExecContext(
		ctx,
		`UPDATE job_applications SET company = ?, title = ?, url = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		rec.Company,
		rec.Title,
		rec.URL,
		strings.ToLower(rec.Status.String()),
		rec.ID,
	)
	if err != nil {
		tx.Rollback()
		return time.Time{}, err
	}
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO job_application_status_histories (job_application_id, status)
                SELECT ?, ?
                WHERE NOT EXISTS (
                    SELECT 1
                    FROM (
                             SELECT status
                             FROM job_application_status_histories
                             WHERE job_application_id = ?
                             ORDER BY created_at DESC
                             LIMIT 1
                         )
                    WHERE status = ?
                )`,
		rec.ID,
		strings.ToLower(rec.Status.String()),
		rec.ID,
		strings.ToLower(rec.Status.String()),
	)
	if err != nil {
		tx.Rollback()
		return time.Time{}, err
	}
	row := tx.QueryRowContext(
		ctx,
		`SELECT updated_at FROM job_applications WHERE id = ?`,
		rec.ID,
	)
	var updatedAt time.Time
	if err = row.Scan(&updatedAt); err != nil {
		tx.Rollback()
		return time.Time{}, err
	}
	return updatedAt, tx.Commit()
}

func (s *JobApplicationStore) Delete(ctx context.Context, id int) error {
	_, err := s.Database.DB().ExecContext(
		ctx,
		`DELETE FROM job_applications WHERE id = ?`,
		id,
	)
	return err
}
