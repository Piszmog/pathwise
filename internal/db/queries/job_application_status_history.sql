-- name: GetLatestJobApplicationStatusHistoryByID :one
SELECT
	h.created_at, h.status, h.id, h.job_application_id
FROM 
	job_application_status_histories h
WHERE
	h.job_application_id = ?
ORDER BY h.created_at DESC
LIMIT 1;

-- name: GetJobApplicationStatusHistoriesByJobApplicationID :many
SELECT
	h.created_at, h.status, h.id, h.job_application_id
FROM 
	job_application_status_histories h
WHERE
	h.job_application_id = ?
ORDER BY h.created_at DESC;

-- name: InsertJobApplicationStatusHistory :exec
INSERT INTO job_application_status_histories (job_application_id) 
VALUES (?);

-- name: InsertJobApplicationStatusHistoryWithStatus :exec
INSERT INTO job_application_status_histories (status, job_application_id)
VALUES (?, ?);

-- name: CountJobApplicationStatusHistoriesByJobApplicationID :one
SELECT COUNT(*) FROM job_application_status_histories WHERE job_application_id = ?;

-- name: GetJobApplicationStatusHistoryByJobApplicationIDAndUserID :many
SELECT
	h.created_at, h.status, h.id, h.job_application_id
FROM 
	job_application_status_histories h
JOIN job_applications ja ON h.job_application_id = ja.id
WHERE
	h.job_application_id = ? AND ja.user_id = ?;

-- name: GetAllJobApplicationStatusHistoryByUserID :many
SELECT
	h.created_at, h.status, h.id, h.job_application_id
FROM 
	job_application_status_histories h
JOIN job_applications ja ON h.job_application_id = ja.id
WHERE
	ja.user_id = ?;
