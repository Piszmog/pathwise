-- name: GetStatusTransitionsForUser :many
SELECT
  h1.status AS from_status,
  h2.status AS to_status,
  COUNT(*) AS transition_count
FROM
  job_application_status_histories h1
  JOIN job_application_status_histories h2 ON h1.job_application_id = h2.job_application_id
  JOIN job_applications ja ON h1.job_application_id = ja.id
WHERE
  ja.user_id = ?
  AND ja.archived = false
  AND h2.created_at > h1.created_at
  AND h2.id = (
    SELECT
      MIN(h3.id)
    FROM
      job_application_status_histories h3
    WHERE
      h3.job_application_id = h1.job_application_id
      AND h3.created_at > h1.created_at
  )
GROUP BY
  h1.status,
  h2.status
ORDER BY
  transition_count DESC;

-- name: GetCurrentStatusCounts :many
SELECT
  STATUS,
  COUNT(*) AS count
FROM
  job_applications
WHERE
  user_id = ?
  AND archived = false
GROUP BY
  STATUS
ORDER BY
  count DESC;

-- name: GetAppliedOnlyCount :one
SELECT
  COUNT(*) AS count
FROM
  job_applications ja
WHERE
  ja.user_id = ?
  AND ja.archived = false
  AND ja.status = 'applied'
  AND NOT EXISTS (
    SELECT 1
    FROM job_application_status_histories h
    WHERE h.job_application_id = ja.id
      AND h.status != 'applied'
  );
