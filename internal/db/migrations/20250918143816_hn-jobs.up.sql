CREATE TABLE IF NOT EXISTS hn_who_is_hirings (
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  title TEXT NOT NULL,
  id INTEGER PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS hn_jobs (
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  company TEXT NOT NULL,
  description TEXT NOT NULL,
  location TEXT,
  role_type TEXT,
  hybrid INTEGER NOT NULL DEFAULT 0,
  remote INTEGER NOT NULL DEFAULT 0,
  id INTEGER PRIMARY KEY,
  hn_who_is_hiring_id INTEGER NOT NULL,
  FOREIGN KEY (hn_who_is_hiring_id) REFERENCES hn_who_is_hirings (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS hn_jobs_hn_who_is_hiring_id_idx ON hn_jobs (hn_who_is_hiring_id);
