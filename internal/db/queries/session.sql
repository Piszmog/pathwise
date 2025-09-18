-- name: InsertSession :exec
INSERT INTO
  sessions (
    expires_at,
    user_agent,
    token,
    ip_address,
    user_id
  )
VALUES
  (?, ?, ?, ?, ?);

-- name: GetSessionByToken :one
SELECT
  created_at,
  expires_at,
  token,
  user_id
FROM
  sessions
WHERE
  token = ?;

-- name: DeleteSessionByUserID :exec
DELETE FROM sessions
WHERE
  user_id = ?;

-- name: DeleteOldUserSessions :exec
DELETE FROM sessions
WHERE
  expires_at < CURRENT_TIMESTAMP
  AND user_id = ?;

-- name: DeleteSessionByToken :exec
DELETE FROM sessions
WHERE
  token = ?;

-- name: UpdateSessionExpiresAt :exec
UPDATE sessions
SET
  expires_at = ?
WHERE
  token = ?
  AND updated_at = CURRENT_TIMESTAMP;
