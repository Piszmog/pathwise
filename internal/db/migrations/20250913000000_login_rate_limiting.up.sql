CREATE TABLE IF NOT EXISTS failed_login_attempts (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS failed_login_attempts_email_time_idx ON failed_login_attempts(email, created_at);