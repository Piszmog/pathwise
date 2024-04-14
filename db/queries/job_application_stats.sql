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

-- name: IncrementNewJobApplicationStat :exec
UPDATE job_application_stats
SET
	total_applications = total_applications + 1,
	total_companies = total_companies + ?,
	total_applied = total_applied + 1
WHERE
	user_id = ?;

-- name: UpdateJobApplicationStat :exec
UPDATE job_application_stats
SET
	total_companies = total_companies + ?,
	average_time_to_hear_back = (average_time_to_hear_back + ?)/2,
	total_accepted = total_accepted + ?,
	total_applied = total_applied + ?,
	total_canceled = total_canceled + ?,
	total_declined = total_declined + ?,
	total_interviewing = total_interviewing + ?,
	total_offers = total_offers + ?,
	total_rejected = total_rejected + ?,
	total_watching = total_watching + ?,
	total_widthdrawn = total_widthdrawn + ?,
	updated_at = CURRENT_TIMESTAMP
WHERE
	user_id = ?;

-- name: InsertNewJobApplicationStat :exec
INSERT INTO job_application_stats (user_id) VALUES (?);
