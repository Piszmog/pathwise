CREATE TABLE IF NOT EXISTS users (
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	email TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	id INTEGER PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS sessions (
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	expires_at DATETIME NOT NULL,
	token TEXT NOT NULL UNIQUE,
	user_agent TEXT NOT NULL,
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS job_applications (
	applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	company TEXT NOT NULL,
	title TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'applied',
	url TEXT,
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS job_applications_user_id_idx ON job_applications(user_id);

CREATE INDEX IF NOT EXISTS job_applications_stats_idx ON job_applications(user_id, id, company, status, applied_at);

CREATE INDEX IF NOT EXISTS job_applications_user_id_updated_at_idx ON job_applications(user_id, updated_at);

CREATE TABLE IF NOT EXISTS job_application_notes (
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	note TEXT NOT NULL,
	id INTEGER PRIMARY KEY,
	job_application_id INTEGER NOT NULL,
	FOREIGN KEY (job_application_id) REFERENCES job_applications(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS job_application_status_histories (
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	status TEXT NOT NULL DEFAULT 'applied',
	id INTEGER PRIMARY KEY,
	job_application_id INTEGER NOT NULL,
	FOREIGN KEY (job_application_id) REFERENCES job_applications(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS status_job_application_id_created_at_idx ON job_application_status_histories(status, job_application_id, created_at);
