CREATE TABLE IF NOT EXISTS job_application_stats (
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	average_time_to_hear_back INTEGER NOT NULL DEFAULT 0,
	total_applications INTEGER NOT NULL DEFAULT 0,
	total_companies INTEGER NOT NULL DEFAULT 0,
	total_accepted INTEGER NOT NULL DEFAULT 0,
	total_applied INTEGER NOT NULL DEFAULT 0,
	total_canceled INTEGER NOT NULL DEFAULT 0,
	total_declined INTEGER NOT NULL DEFAULT 0,
	total_interviewing INTEGER NOT NULL DEFAULT 0,
	total_offers INTEGER NOT NULL DEFAULT 0,
	total_rejected INTEGER NOT NULL DEFAULT 0,
	total_watching INTEGER NOT NULL DEFAULT 0,
	total_widthdrawn INTEGER NOT NULL DEFAULT 0,
	user_id INTEGER NOT NULL,
	id INTEGER PRIMARY KEY,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS stats_user_id_idx ON job_application_stats(user_id);

INSERT INTO job_application_stats (user_id, total_applications, total_companies, average_time_to_hear_back, total_accepted, total_applied, total_canceled, total_declined, total_interviewing, total_offers, total_rejected, total_watching, total_widthdrawn)
SELECT 
	user_id,
	COUNT(*) AS total_applications,
	COUNT(DISTINCT company) AS total_companies,
	ROUND(IFNULL(AVG(JULIANDAY(jh_created_at) - JULIANDAY(applied_at)), 0), 0) AS average_time_to_hear_back,
	COUNT(CASE WHEN status = 'accepted' THEN 1 END) AS total_accepted,
	COUNT(CASE WHEN status = 'applied' THEN 1 END) AS total_applied,
	COUNT(CASE WHEN status = 'canceled' THEN 1 END) AS total_canceled,
	COUNT(CASE WHEN status = 'declined' THEN 1 END) AS total_declined,
	COUNT(CASE WHEN status = 'interviewing' THEN 1 END) AS total_interviewing,
	COUNT(CASE WHEN status = 'offer' THEN 1 END) AS total_offers,
	COUNT(CASE WHEN status = 'rejected' THEN 1 END) AS total_rejected,
	COUNT(CASE WHEN status = 'watching' THEN 1 END) AS total_watching,
	COUNT(CASE WHEN status = 'withdrawn' THEN 1 END) AS total_widthdrawn
FROM job_applications
LEFT JOIN (
	SELECT
		job_application_id,
		MIN(created_at) AS jh_created_at
	FROM
		job_application_status_histories
	WHERE
		status IN ('interviewing', 'rejected', 'cancelled', 'closed')
	GROUP BY job_application_id 
) jh ON job_applications.id = jh.job_application_id
GROUP BY user_id;
