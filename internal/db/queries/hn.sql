-- name: ExistsHNStory :one
SELECT
  1
FROM
  hn_stories
WHERE
  id = ?
LIMIT
  1;

-- name: InsertHNStory :exec
INSERT INTO
  hn_stories (posted_at, title, id)
VALUES
  (?, ?, ?);

-- name: ExistsHNComment :one
SELECT
  1
FROM
  hn_comments
WHERE
  id = ?
LIMIT
  1;

-- name: InsertHNComment :exec
INSERT INTO
  hn_comments (commented_at, value, id, hn_story_id)
VALUES
  (?, ?, ?, ?);

-- name: UpdateHNComments :exec
UPDATE hn_comments
SET
  updated_at = CURRENT_TIMESTAMP,
  status = ?
WHERE
  id IN (sqlc.slice ('ids'));

-- name: GetQueuedHNComments :many
SELECT
  id
FROM
  hn_comments
WHERE
  status IN ('queued', 'in_progress', 'failed');

-- name: GetHNCommentValues :many
SELECT
  id,
  value
FROM
  hn_comments
WHERE
  id IN (sqlc.slice ('ids'));

-- name: InsertHNJob :exec
INSERT INTO
  hn_jobs (
    company,
    company_description,
    title,
    id,
    company_url,
    contact_email,
    description,
    role_type,
    location,
    salary,
    equity,
    is_hybrid,
    is_remote,
    hn_comment_id
  )
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertHNTechStack :exec
INSERT INTO
  hn_job_tech_stacks (hn_job_id, value)
VALUES
  (?, ?);

-- name: GetHNJobs :many
SELECT
  hn_jobs.id,
  hn_jobs.company,
  hn_jobs.title,
  hn_jobs.role_type,
  hn_jobs.location,
  hn_jobs.salary,
  hn_jobs.is_hybrid,
  hn_jobs.is_remote
FROM
  hn_jobs
  LEFT JOIN hn_comments ON hn_comments.id = hn_jobs.hn_comment_id
ORDER BY
  hn_comments.commented_at DESC;

-- name: GetHNJobByID :one
SELECT
  id,
  company,
  company_description,
  title,
  company_url,
  contact_email,
  description,
  role_type,
  location,
  salary,
  equity,
  is_hybrid,
  is_remote,
  created_at,
  hn_comment_id
FROM
  hn_jobs
WHERE
  id = ?;

-- name: GetHNJobTechStacks :many
SELECT
  value
FROM
  hn_job_tech_stacks
WHERE
  hn_job_id = ?;

-- name: GetHNJobsPaginated :many
SELECT
  hn_jobs.id,
  hn_jobs.company,
  hn_jobs.title,
  hn_jobs.role_type,
  hn_jobs.location,
  hn_jobs.salary,
  hn_jobs.is_hybrid,
  hn_jobs.is_remote
FROM
  hn_jobs
  LEFT JOIN hn_comments ON hn_comments.id = hn_jobs.hn_comment_id
ORDER BY
  hn_comments.commented_at DESC
LIMIT
  ?
OFFSET
  ?;

-- name: GetHNJobsFiltered :many
SELECT
  hn_jobs.id,
  hn_jobs.company,
  hn_jobs.title,
  hn_jobs.role_type,
  hn_jobs.location,
  hn_jobs.salary,
  hn_jobs.is_hybrid,
  hn_jobs.is_remote
FROM
  hn_jobs
  LEFT JOIN hn_comments ON hn_comments.id = hn_jobs.hn_comment_id
WHERE
  (
    ? = -1
    OR hn_jobs.is_remote = ?
  )
  AND (
    ? = -1
    OR hn_jobs.is_hybrid = ?
  )
  AND (
    ? = ''
    OR hn_jobs.id IN (
      SELECT
        hn_job_id
      FROM
        hn_job_tech_stacks
      WHERE
        hn_job_tech_stacks.value = ?
    )
  )
ORDER BY
  hn_comments.commented_at DESC
LIMIT
  ?
OFFSET
  ?;
