-- name: ExistsHNStory :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      hn_stories
    WHERE
      id = ?
  );

-- name: InsertHNStory :exec
INSERT INTO
  hn_stories (posted_at, title, id)
VALUES
  (?, ?, ?);

-- name: ExistsHNComment :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      hn_comments
    WHERE
      id = ?
  );

-- name: InsertHNComment :exec
INSERT INTO
  hn_comments (commented_at, value, id, hn_story_id)
VALUES
  (?, ?, ?, ?);

-- name: UpdateHNComment :exec
UPDATE hn_comments
SET
  updated_at = CURRENT_TIMESTAMP,
  status = ?
WHERE
  id = ?;

-- name: GetQueuedHNComments :many
SELECT
  id
FROM
  hn_comments
WHERE
  status IN ('queued', 'in_progress');

-- name: GetHNCommentValue :one
SELECT
  value
FROM
  hn_comments
WHERE
  id = ?;

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
