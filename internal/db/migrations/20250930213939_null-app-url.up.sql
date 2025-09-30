CREATE TABLE IF NOT EXISTS new_job_applications (
  applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  company TEXT NOT NULL,
  title TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'applied',
  url TEXT,
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL,
  archived INTEGER NOT NULL DEFAULT 0,
  salary_min INTEGER,
  salary_max INTEGER,
  salary_currency TEXT,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

INSERT INTO
  new_job_applications
SELECT
  applied_at,
  updated_at,
  created_at,
  company,
  title,
  status,
  url,
  id,
  user_id,
  archived,
  salary_min,
  salary_max,
  salary_currency
FROM
  job_applications;

DROP TABLE job_applications;

ALTER TABLE new_job_applications
RENAME TO job_applications;

CREATE INDEX IF NOT EXISTS job_applications_user_id_idx ON job_applications (user_id);

CREATE INDEX IF NOT EXISTS job_applications_stats_idx ON job_applications (user_id, id, company, status, applied_at);

CREATE INDEX IF NOT EXISTS job_applications_user_id_updated_at_idx ON job_applications (user_id, updated_at);
