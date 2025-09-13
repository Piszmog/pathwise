-- name: InsertFailedLoginAttempt :exec
INSERT INTO failed_login_attempts (id, email, ip_address)
VALUES (?, ?, ?);

-- name: GetRecentFailedAttempts :one
SELECT COUNT(*) as count
FROM failed_login_attempts 
WHERE email = ? AND created_at > ?;

-- name: GetLastFailedAttempt :one
SELECT created_at
FROM failed_login_attempts
WHERE email = ?
ORDER BY created_at DESC
LIMIT 1;