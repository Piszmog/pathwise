DROP INDEX IF EXISTS session_token_idx;

CREATE INDEX IF NOT EXISTS job_applications_stats_idx ON job_applications(user_id, id, company, status, applied_at);

DROP INDEX IF EXISTS job_application_notes_job_application_id_idx;

DROP INDEX IF EXISTS job_application_status_histories_job_application_id_created_at_idx;

CREATE INDEX IF NOT EXISTS status_job_application_id_created_at_idx ON job_application_status_histories(status, job_application_id, created_at);
