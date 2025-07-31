CREATE TABLE IF NOT EXISTS new_job_applications (
	applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	company TEXT NOT NULL,
	title TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'applied',
	url TEXT NOT NULL,
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO new_job_applications (
	applied_at,
	updated_at,
	created_at,
	company,
	title,
	status,
	url,
	id,
	user_id
) SELECT
	applied_at,
	updated_at,
	created_at,
	company,
	title,
	status,
	url,
	id,
	user_id
FROM
	job_applications;

DROP TABLE job_applications;

ALTER TABLE new_job_applications RENAME TO job_applications;
