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
