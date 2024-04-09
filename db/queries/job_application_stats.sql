-- name: GetJobApplicationStat :one
SELECT 
	average_time_to_hear_back,
	total_applications,
	total_companies,
	total_interviewing,
	total_rejected
FROM
	job_application_stats
WHERE
	user_id = ?;

-- name: IncrementNewJobApplication :exec
UPDATE job_application_stats
SET
	total_applications = total_applications + 1,
	total_companies = ?,
	total_applied = total_applied + 1
WHERE
	user_id = ?;

-- name: CreateNewJobApplicationStat :exec
INSERT INTO job_application_stats (user_id) VALUES (?);
