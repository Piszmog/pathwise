-- name: GetJobApplicationByID :one
SELECT
	j.applied_at, j.updated_at, j.company, j.title, j.status, j.url, j.id, j.user_id, j.archived
FROM 
	job_applications j
WHERE
	j.id = ?;

-- name: GetJobApplicationByIDAndUserID :one
SELECT
	j.applied_at, j.updated_at, j.company, j.title, j.status, j.url, j.id, j.user_id
FROM
	job_applications j
WHERE
	j.id = ? AND j.user_id = ?;

-- name: GetJobApplicationsByUserID :many 
SELECT
	j.applied_at, j.updated_at, j.company, j.title, j.status, j.url, j.id
FROM 
	job_applications j
WHERE
	j.user_id = ? AND j.archived = ?
ORDER BY j.updated_at DESC
LIMIT ? OFFSET ?;

-- name: CountJobApplicationsByUserID :one
SELECT COUNT(*) FROM job_applications WHERE user_id = ? AND archived = ?;

-- name: GetJobApplicationsByUserIDAndCompany :many 
SELECT
	j.applied_at, j.updated_at, j.company, j.title, j.status, j.url, j.id
FROM 
	job_applications j
WHERE
	j.company LIKE ? AND j.user_id = ? AND j.archived = ?
ORDER BY j.updated_at DESC
LIMIT ? OFFSET ?;

-- name: CountJobApplicationsByUserIDAndCompany :one
SELECT COUNT(*) FROM job_applications WHERE company LIKE ? AND user_id = ? AND archived = ?;

-- name: GetJobApplicationsByUserIDAndStatus :many 
SELECT
	j.applied_at, j.updated_at, j.company, j.title, j.status, j.url, j.id
FROM 
	job_applications j
WHERE
	j.status = ? AND j.user_id = ? AND j.archived = ?
ORDER BY j.updated_at DESC
LIMIT ? OFFSET ?;

-- name: CountJobApplicationsByUserIDAndStatus :one
SELECT COUNT(*) FROM job_applications WHERE status = ? AND user_id = ? AND archived = ?;

-- name: GetJobApplicationsByUserIDAndCompanyAndStatus :many 
SELECT
	j.applied_at, j.updated_at, j.company, j.title, j.status, j.url, j.id
FROM 
	job_applications j
WHERE
	j.company LIKE ? AND j.status = ? AND j.user_id = ? AND j.archived = ?
ORDER BY j.updated_at DESC
LIMIT ? OFFSET ?;

-- name: CountJobApplicationsByUserIDAndCompanyAndStatus :one
SELECT COUNT(*) FROM job_applications WHERE company = ? AND status = ? AND user_id = ? AND archived = ?;

-- name: GetJobApplicationUpdatedAt :one
SELECT updated_at FROM job_applications WHERE id = ?;

-- name: CountJobApplicationCompany :one
SELECT COUNT(*) FROM job_applications WHERE company = ? AND user_id = ? AND archived = ?;

-- name: InsertJobApplication :one 
INSERT INTO job_applications (company, title, url, user_id) 
VALUES (?, ?, ?, ?)
RETURNING id;

-- name: UpdateJobApplication :exec
UPDATE job_applications 
	SET company = ?, 
		title = ?, 
		status = ?, 
		url = ?,
		updated_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: ArchiveJobApplications :exec
UPDATE job_applications
	SET archived = 1,
		updated_at = CURRENT_TIMESTAMP
WHERE user_id = ?
AND applied_at <= ?;

-- name: CountJobApplicationsForStats :one
SELECT
	COUNT(*)
FROM 
	job_applications j
JOIN job_application_status_histories h 
	ON h.job_application_id = j.id
WHERE
	j.user_id = ? AND j.archived = 0;

-- name: GetJobApplicationsForStats :many 
SELECT
	j.applied_at, j.status, MIN(h.created_at) AS heard_back_at
FROM 
	job_applications j
JOIN job_application_status_histories h 
	ON h.job_application_id = j.id
WHERE
	j.user_id = ? AND j.archived = 0
GROUP BY
	j.id;

-- name: CountJobApplicationCompanies :one
SELECT
	COUNT(DISTINCT company)
FROM
	job_applications j
WHERE
	j.user_id = ? AND j.archived = ?;
