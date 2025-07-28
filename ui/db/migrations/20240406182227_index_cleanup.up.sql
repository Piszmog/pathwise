CREATE INDEX IF NOT EXISTS sessions_token_idx ON sessions(token);

DROP INDEX IF EXISTS job_applications_stats_idx;

CREATE INDEX IF NOT EXISTS job_application_notes_job_application_id_idx ON job_application_notes(job_application_id);

DROP INDEX IF EXISTS status_job_application_id_created_at_idx;

CREATE INDEX IF NOT EXISTS job_application_status_histories_job_application_id_created_at_idx ON job_application_status_histories(job_application_id, created_at);

CREATE INDEX IF NOT EXISTS job_applications_user_id_status_updated_at_idx ON job_applications(user_id, status, updated_at);

CREATE INDEX IF NOT EXISTS job_applications_user_id_company_updated_at_idx ON job_applications(user_id, company, updated_at);
