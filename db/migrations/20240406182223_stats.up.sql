CREATE TABLE IF NOT EXISTS job_application_stats (
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	average_time_to_hear_back INTEGER,
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
