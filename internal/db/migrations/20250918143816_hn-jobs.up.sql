CREATE TABLE IF NOT EXISTS hn_stories (
  posted_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  title TEXT NOT NULL,
  id INTEGER PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS hn_comments (
  commented_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status TEXT CHECK (
    status IN ('queued', 'in_progress', 'completed', 'failed')
  ) NOT NULL DEFAULT 'queued',
  value TEXT NOT NULL,
  id INTEGER PRIMARY KEY,
  hn_story_id INTEGER NOT NULL,
  FOREIGN KEY (hn_story_id) REFERENCES hn_stories (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS hn_comments_story_id_idx ON hn_comments (hn_story_id);

CREATE INDEX IF NOT EXISTS hn_comments_status_idx ON hn_comments (status);

CREATE TABLE IF NOT EXISTS hn_jobs (
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  company TEXT NOT NULL,
  company_description TEXT NOT NULL,
  title TEXT NOT NULL,
  id TEXT PRIMARY KEY,
  company_url TEXT,
  contact_email TEXT,
  description TEXT,
  role_type TEXT,
  location TEXT,
  salary TEXT,
  equity TEXT,
  is_hybrid INTEGER NOT NULL DEFAULT 0,
  is_remote INTEGER NOT NULL DEFAULT 0,
  hn_comment_id INTEGER NOT NULL,
  FOREIGN KEY (hn_comment_id) REFERENCES hn_comments (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS hn_jobs_comment_id_idx ON hn_jobs (hn_comment_id);

CREATE INDEX IF NOT EXISTS hn_jobs_is_hybrid_idx ON hn_jobs (is_hybrid);

CREATE INDEX IF NOT EXISTS hn_jobs_is_remote_idx ON hn_jobs (is_remote);

CREATE INDEX IF NOT EXISTS hn_jobs_title_idx ON hn_jobs (title);

CREATE TABLE IF NOT EXISTS hn_job_tech_stacks (
  hn_job_id TEXT NOT NULL,
  value TEXT NOT NULL,
  PRIMARY KEY (hn_job_id, value),
  FOREIGN KEY (hn_job_id) REFERENCES hn_jobs (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS hn_job_tech_stacks_hn_job_id_idx ON hn_job_tech_stacks (hn_job_id);

CREATE INDEX IF NOT EXISTS hn_job_tech_stacks_value_idx ON hn_job_tech_stacks (value);
