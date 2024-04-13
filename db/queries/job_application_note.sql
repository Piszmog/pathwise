-- name: GetJobApplicationNotesByJobApplicationID :many
SELECT
	n.created_at, n.note, n.job_application_id, n.id
FROM 
	job_application_notes n
WHERE
	n.job_application_id = ?
ORDER BY n.created_at DESC;

-- name: InsertJobApplicationNote :one
INSERT INTO job_application_notes (note, job_application_id) 
VALUES (?, ?)
RETURNING created_at, note, job_application_id, id;

-- name: GetJobApplicationNoteByID :many
SELECT
	 n.created_at, n.note, n.id, n.job_application_id
FROM 
	job_application_notes n
WHERE
	n.id = ?;
