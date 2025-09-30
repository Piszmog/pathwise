DROP INDEX IF EXISTS hn_jobs_title_idx;

DROP INDEX IF EXISTS hn_jobs_location_idx;

CREATE INDEX IF NOT EXISTS hn_jobs_work_type_date_idx ON hn_jobs(is_remote, is_hybrid, hn_comment_id);