package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
)

type JobApplicationStore struct {
	Database db.Database
}

func (s *JobApplicationStore) GetByID(ctx context.Context, id int) (types.JobApplication, error) {
	return scanJobApplication(s.Database.DB().QueryRowContext(ctx, jobGetByIDQuery, id))
}

const jobGetByIDQuery = `
SELECT
	j.id, j.company, j.title, j.url, j.status, j.applied_at, j.updated_at, j.user_id
FROM 
	job_applications j
WHERE
	j.id = ?
`

func (s *JobApplicationStore) GetByIDAndUserID(ctx context.Context, id, userID int) (types.JobApplication, error) {
	return scanJobApplication(s.Database.DB().QueryRowContext(ctx, jobGetByIDAndUserIDQuery, id, userID))
}

const jobGetByIDAndUserIDQuery = `
SELECT
	j.id, j.company, j.title, j.url, j.status, j.applied_at, j.updated_at, j.user_id
FROM
	job_applications j
WHERE
	j.id = ? AND j.user_id = ?
`

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
		&job.UserID,
	)
	job.Status = types.ToJobApplicationStatus(status)
	return job, err
}

func (s *JobApplicationStore) Get(ctx context.Context, userID int, opts LimitOpts) ([]types.JobApplication, int, error) {
	tx, err := s.Database.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}

	rows, err := tx.QueryContext(ctx, jobGetLimitQuery, userID, opts.PerPage, opts.Page*opts.PerPage)
	if err != nil {
		_ = tx.Rollback()
		return nil, 0, err
	}
	jobs, err := scanJobApplications(rows)
	if err != nil {
		_ = tx.Rollback()
		return nil, 0, err
	}
	row := tx.QueryRowContext(ctx, jobCountQuery, userID)
	var total int
	if err = row.Scan(&total); err != nil {
		_ = tx.Rollback()
		return nil, 0, err
	}
	return jobs, total, tx.Commit()
}

const jobGetLimitQuery = `
SELECT
	j.id, j.company, j.title, j.url, j.status, j.applied_at, j.updated_at
FROM 
	job_applications j
WHERE
	j.user_id = ?
ORDER BY j.updated_at DESC
LIMIT ? OFFSET ?
`

const jobCountQuery = `SELECT COUNT(*) FROM job_applications WHERE user_id = ?`

func (s *JobApplicationStore) Filter(ctx context.Context, opts LimitOpts, userID int, company string, status types.JobApplicationStatus) ([]types.JobApplication, int, error) {
	query := `SELECT
		    j.id, j.company, j.title, j.url, j.status, j.applied_at, j.updated_at
		FROM 
		    job_applications j
		WHERE
		    j.user_id = ?`
	totalQuery := `SELECT COUNT(*) FROM job_applications WHERE user_id = ?`
	if company != "" || status != "" {
		if company != "" {
			query += ` AND j.company LIKE ?`
			totalQuery += ` AND company LIKE ?`
		}
		if status != "" {
			query += ` AND j.status LIKE ?`
			totalQuery += ` AND status LIKE ?`
		}
	}
	query += ` ORDER BY j.updated_at DESC LIMIT ? OFFSET ?`

	queryArgs := []interface{}{userID}
	totalQueryArgs := []interface{}{userID}
	if company != "" {
		queryArgs = append(queryArgs, "%"+company+"%")
		totalQueryArgs = append(totalQueryArgs, "%"+company+"%")
	}
	if status != "" {
		queryArgs = append(queryArgs, "%"+status.String()+"%")
		totalQueryArgs = append(totalQueryArgs, "%"+status.String()+"%")
	}
	queryArgs = append(queryArgs, opts.PerPage, opts.Page*opts.PerPage)

	tx, err := s.Database.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}

	rows, err := tx.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		_ = tx.Rollback()
		return nil, 0, err
	}
	jobs, err := scanJobApplications(rows)
	if err != nil {
		_ = tx.Rollback()
		return nil, 0, err
	}
	row := tx.QueryRowContext(ctx, totalQuery, totalQueryArgs...)
	var total int
	if err = row.Scan(&total); err != nil {
		_ = tx.Rollback()
		return nil, 0, err
	}
	return jobs, total, tx.Commit()
}

func scanJobApplications(rows *sql.Rows) ([]types.JobApplication, error) {
	defer rows.Close()
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
	res, err := tx.ExecContext(ctx, jobInsertQuery, rec.Company, rec.Title, rec.URL, rec.UserID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err = tx.ExecContext(ctx, jobInsertStatusHistory, id); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

const jobInsertQuery = `INSERT INTO job_applications (company, title, url, user_id) VALUES (?, ?, ?, ?)`

const jobInsertStatusHistory = `INSERT INTO job_application_status_histories (job_application_id) VALUES (?)`

func (s *JobApplicationStore) Update(ctx context.Context, rec types.JobApplication) (time.Time, error) {
	tx, err := s.Database.DB().BeginTx(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}
	statusString := rec.Status.String()
	if _, err = tx.ExecContext(ctx, jobUpdateQuery, rec.Company, rec.Title, rec.URL, statusString, rec.ID); err != nil {
		_ = tx.Rollback()
		return time.Time{}, err
	}
	if _, err = tx.ExecContext(ctx, jobInsertStatusHistoryQuery, rec.ID, statusString, rec.ID, statusString); err != nil {
		_ = tx.Rollback()
		return time.Time{}, err
	}
	row := tx.QueryRowContext(ctx, jobGetUpdatedAtQuery, rec.ID)
	var updatedAt time.Time
	if err = row.Scan(&updatedAt); err != nil {
		_ = tx.Rollback()
		return time.Time{}, err
	}
	return updatedAt, tx.Commit()
}

const jobUpdateQuery = `UPDATE job_applications SET company = ?, title = ?, url = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

const jobInsertStatusHistoryQuery = `
INSERT INTO job_application_status_histories (job_application_id, status)
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
)
`

const jobGetUpdatedAtQuery = `SELECT updated_at FROM job_applications WHERE id = ?`
