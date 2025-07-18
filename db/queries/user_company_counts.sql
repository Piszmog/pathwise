-- name: GetCompanyCount :one
SELECT count FROM user_company_counts 
WHERE user_id = ? AND company = ? AND archived = ?;

-- name: GetCompanyCountsByUserID :many
SELECT company, count FROM user_company_counts 
WHERE user_id = ? AND archived = ?
ORDER BY count DESC;

-- name: UpsertCompanyCount :exec
INSERT INTO user_company_counts (user_id, company, archived, count, updated_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(user_id, company, archived) 
DO UPDATE SET 
    count = excluded.count,
    updated_at = CURRENT_TIMESTAMP;

-- name: IncrementCompanyCount :exec
INSERT INTO user_company_counts (user_id, company, archived, count, updated_at)
VALUES (?, ?, ?, 1, CURRENT_TIMESTAMP)
ON CONFLICT(user_id, company, archived) 
DO UPDATE SET 
    count = count + 1,
    updated_at = CURRENT_TIMESTAMP;

-- name: DecrementCompanyCount :exec
UPDATE user_company_counts 
SET count = CASE WHEN count > 0 THEN count - 1 ELSE 0 END,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND company = ? AND archived = ?;

-- name: DeleteCompanyCountsForUser :exec
DELETE FROM user_company_counts WHERE user_id = ?;