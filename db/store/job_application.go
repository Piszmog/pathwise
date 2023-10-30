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

func (s *JobApplicationStore) Get(ctx context.Context, opts LimitOpts) ([]types.JobApplication, int, error) {
	tx, err := s.Database.DB().BeginTx(ctx, nil)
	rows, err := tx.QueryContext(
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
		tx.Rollback()
		return nil, 0, err
	}
	defer rows.Close()
	jobs, err := scanJobApplications(rows)
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}
	row := tx.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM job_applications`,
	)
	var total int
	if err = row.Scan(&total); err != nil {
		tx.Rollback()
		return nil, 0, err
	}
	return jobs, total, tx.Commit()
}

func (s *JobApplicationStore) Filter(ctx context.Context, opts LimitOpts, company string, status string) ([]types.JobApplication, int, error) {
	query := `SELECT
		    j.id, j.company, j.title, j.url, j.status, j.applied_at, j.updated_at
		FROM 
		    job_applications j`
	totalQuery := `SELECT COUNT(*) FROM job_applications`
	if company != "" || status != "" {
		query += ` WHERE`
		totalQuery += ` WHERE`
		if company != "" {
			query += ` j.company LIKE ?`
			totalQuery += ` company LIKE ?`
		}
		if status != "" {
			if company != "" {
				query += ` AND`
				totalQuery += ` AND`
			}
			query += ` j.status LIKE ?`
			totalQuery += ` status LIKE ?`
		}
	}
	query += ` ORDER BY j.updated_at DESC LIMIT ? OFFSET ?`
	var queryArgs []interface{}
	var totalQueryArgs []interface{}
	if company != "" {
		queryArgs = append(queryArgs, "%"+company+"%")
		totalQueryArgs = append(totalQueryArgs, "%"+company+"%")
	}
	if status != "" {
		queryArgs = append(queryArgs, "%"+status+"%")
		totalQueryArgs = append(totalQueryArgs, "%"+status+"%")
	}
	queryArgs = append(queryArgs, opts.PerPage, opts.Page*opts.PerPage)

	tx, err := s.Database.DB().BeginTx(ctx, nil)
	rows, err := tx.QueryContext(
		ctx,
		query,
		queryArgs...,
	)
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}
	defer rows.Close()
	jobs, err := scanJobApplications(rows)
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}
	row := tx.QueryRowContext(
		ctx,
		totalQuery,
		totalQueryArgs...,
	)
	var total int
	if err = row.Scan(&total); err != nil {
		tx.Rollback()
		return nil, 0, err
	}
	return jobs, total, tx.Commit()
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
