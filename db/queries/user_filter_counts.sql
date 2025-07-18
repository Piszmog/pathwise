-- name: GetFilterCount :one
SELECT count FROM user_filter_counts 
WHERE user_id = ? AND archived = ? AND 
      (company IS NULL OR company = ?) AND 
      (status IS NULL OR status = ?);

-- name: GetFilterCountByUserIDAndArchived :one
SELECT count FROM user_filter_counts 
WHERE user_id = ? AND archived = ? AND company IS NULL AND status IS NULL;

-- name: GetFilterCountByUserIDAndCompany :one
SELECT count FROM user_filter_counts 
WHERE user_id = ? AND archived = ? AND company = ? AND status IS NULL;

-- name: GetFilterCountByUserIDAndStatus :one
SELECT count FROM user_filter_counts 
WHERE user_id = ? AND archived = ? AND company IS NULL AND status = ?;

-- name: GetFilterCountByUserIDAndCompanyAndStatus :one
SELECT count FROM user_filter_counts 
WHERE user_id = ? AND archived = ? AND company = ? AND status = ?;

-- name: UpsertFilterCount :exec
INSERT INTO user_filter_counts (user_id, archived, company, status, count, updated_at)
VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(user_id, archived, company, status) 
DO UPDATE SET 
    count = excluded.count,
    updated_at = CURRENT_TIMESTAMP;

-- name: IncrementFilterCount :exec
INSERT INTO user_filter_counts (user_id, archived, company, status, count, updated_at)
VALUES (?, ?, ?, ?, 1, CURRENT_TIMESTAMP)
ON CONFLICT(user_id, archived, company, status) 
DO UPDATE SET 
    count = count + 1,
    updated_at = CURRENT_TIMESTAMP;

-- name: DecrementFilterCount :exec
UPDATE user_filter_counts 
SET count = CASE WHEN count > 0 THEN count - 1 ELSE 0 END,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND archived = ? AND 
      (company IS NULL OR company = ?) AND 
      (status IS NULL OR status = ?);

-- name: DeleteFilterCountsForUser :exec
DELETE FROM user_filter_counts WHERE user_id = ?;