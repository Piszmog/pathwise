-- name: GetJobApplicationNotesByJobApplicationID :many
SELECT
  n.created_at,
  n.note,
  n.job_application_id,
  n.id
FROM
  job_application_notes n
WHERE
  n.job_application_id = ?
ORDER BY
  n.created_at DESC;

-- name: InsertJobApplicationNote :one
INSERT INTO
  job_application_notes (note, job_application_id)
VALUES
  (?, ?) RETURNING created_at,
  note,
  job_application_id,
  id;

-- name: GetJobApplicationNotesByJobApplicationIDAndUserID :many
SELECT
  n.created_at,
  n.note,
  n.job_application_id,
  n.id
FROM
  job_application_notes n
  JOIN job_applications ja ON n.job_application_id = ja.id
WHERE
  n.job_application_id = ?
  AND ja.user_id = ?;

-- name: GetAllJobApplicationNotesByUserID :many
SELECT
  n.created_at,
  n.note,
  n.job_application_id,
  n.id
FROM
  job_application_notes n
  JOIN job_applications ja ON n.job_application_id = ja.id
WHERE
  ja.user_id = ?;

