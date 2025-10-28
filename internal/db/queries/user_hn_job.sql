-- name: CheckUserHasAddedHNJob :one
SELECT 
    job_application_id
FROM 
    user_hn_jobs
WHERE 
    user_id = ? 
    AND hn_job_id = ?;

-- name: InsertUserHNJob :exec
INSERT INTO user_hn_jobs (
    user_id,
    hn_job_id,
    job_application_id
) VALUES (?, ?, ?);

-- name: DeleteUserHNJob :exec
DELETE FROM user_hn_jobs
WHERE 
    user_id = ? 
    AND hn_job_id = ?;
