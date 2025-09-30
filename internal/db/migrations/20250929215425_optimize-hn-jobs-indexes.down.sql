DROP INDEX IF EXISTS hn_jobs_work_type_date_idx;

CREATE INDEX IF NOT EXISTS hn_jobs_location_idx ON hn_jobs(location);

CREATE INDEX IF NOT EXISTS hn_jobs_title_idx ON hn_jobs(title);