-- name: GetStatusTransitionsForUser :many
SELECT 
    h1.status as from_status,
    h2.status as to_status,
    COUNT(*) as transition_count
FROM job_application_status_histories h1
JOIN job_application_status_histories h2 ON h1.job_application_id = h2.job_application_id
JOIN job_applications ja ON h1.job_application_id = ja.id
WHERE 
    ja.user_id = ? 
    AND ja.archived = false
    AND h2.created_at > h1.created_at
    AND h2.id = (
        SELECT MIN(h3.id) 
        FROM job_application_status_histories h3 
        WHERE h3.job_application_id = h1.job_application_id 
        AND h3.created_at > h1.created_at
    )
GROUP BY h1.status, h2.status
ORDER BY transition_count DESC;

-- name: GetCurrentStatusCounts :many
SELECT status, COUNT(*) as count
FROM job_applications
WHERE user_id = ? AND archived = false
GROUP BY status
ORDER BY count DESC;

-- name: GetAnalyticsStats :one
SELECT 
    COUNT(*) as total_applications,
    COUNT(CASE WHEN status = 'interviewing' THEN 1 END) as total_interviewing,
    COUNT(CASE WHEN status = 'accepted' THEN 1 END) as total_accepted
FROM job_applications
WHERE user_id = ? AND archived = false;
