CREATE TABLE user_hn_jobs (
    user_id INTEGER NOT NULL,
    hn_job_id TEXT NOT NULL,
    job_application_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, hn_job_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (hn_job_id) REFERENCES hn_jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (job_application_id) REFERENCES job_applications(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_hn_jobs_user_id ON user_hn_jobs(user_id);
CREATE INDEX idx_user_hn_jobs_hn_job_id ON user_hn_jobs(hn_job_id);
