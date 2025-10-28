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
  status IN (sqlc.slice ('statuses'));

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
    application_url,
    jobs_url,
    role_type,
    location,
    salary,
    equity,
    is_hybrid,
    is_remote,
    hn_comment_id
  )
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

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
  application_url,
  jobs_url,
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
  hn_jobs.is_remote,
  hn_comments.commented_at
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
  hn_jobs.is_remote,
  hn_comments.commented_at
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

-- name: SearchHNJobs :many
SELECT DISTINCT
  j.id,
  j.title,
  j.company,
  j.location,
  j.is_remote,
  j.is_hybrid,
  hc.commented_at as posted
FROM
  hn_jobs j
  LEFT JOIN hn_job_tech_stacks ts ON j.id = ts.hn_job_id
  LEFT JOIN hn_comments hc ON j.hn_comment_id = hc.id
WHERE
  1 = 1
  AND (
    sqlc.narg ('title') IS NULL
    OR j.title LIKE '%' || sqlc.narg ('title') || '%'
  )
  AND (
    sqlc.narg ('location') IS NULL
    OR j.location LIKE '%' || sqlc.narg ('location') || '%'
  )
  AND (
    (
      sqlc.narg ('is_remote') IS NULL
      OR sqlc.narg ('is_remote') = 0
    )
    OR j.is_remote = 1
  )
  AND (
    (
      sqlc.narg ('is_hybrid') IS NULL
      OR sqlc.narg ('is_hybrid') = 0
    )
    OR j.is_hybrid = 1
  )
  AND (
    sqlc.narg ('keyword') IS NULL
    OR sqlc.narg ('keyword') = ''
    OR j.description LIKE '%' || sqlc.narg ('keyword') || '%'
    OR j.company_description LIKE '%' || sqlc.narg ('keyword') || '%'
  )
  AND (
    sqlc.narg ('tech_stack') IS NULL
    OR LOWER(ts.value) IN (sqlc.narg ('tech_stack'))
  )
ORDER BY
  posted DESC
LIMIT
  sqlc.arg ('limit')
OFFSET
  sqlc.arg ('offset');
