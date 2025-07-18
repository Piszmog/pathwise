-- Add archived tracking to existing stats table
ALTER TABLE job_application_stats ADD COLUMN total_applications_archived INTEGER NOT NULL DEFAULT 0;
ALTER TABLE job_application_stats ADD COLUMN total_companies_archived INTEGER NOT NULL DEFAULT 0;

-- Create user filter counts table for pagination
CREATE TABLE IF NOT EXISTS user_filter_counts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    archived BOOLEAN NOT NULL DEFAULT 0,
    company TEXT,
    status TEXT,
    count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS user_filter_counts_unique_idx 
ON user_filter_counts(user_id, archived, company, status);

CREATE INDEX IF NOT EXISTS user_filter_counts_user_archived_idx 
ON user_filter_counts(user_id, archived);

-- Create company counts table
CREATE TABLE IF NOT EXISTS user_company_counts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    company TEXT NOT NULL,
    archived BOOLEAN NOT NULL DEFAULT 0,
    count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS user_company_counts_unique_idx 
ON user_company_counts(user_id, company, archived);

-- Populate user_filter_counts with existing data
-- Base counts (no filters)
INSERT INTO user_filter_counts (user_id, archived, company, status, count)
SELECT 
    user_id, 
    archived, 
    NULL as company, 
    NULL as status, 
    COUNT(*) as count
FROM job_applications 
GROUP BY user_id, archived;

-- Company filter counts
INSERT INTO user_filter_counts (user_id, archived, company, status, count)
SELECT 
    user_id, 
    archived, 
    company, 
    NULL as status, 
    COUNT(*) as count
FROM job_applications 
GROUP BY user_id, archived, company;

-- Status filter counts
INSERT INTO user_filter_counts (user_id, archived, company, status, count)
SELECT 
    user_id, 
    archived, 
    NULL as company, 
    status, 
    COUNT(*) as count
FROM job_applications 
GROUP BY user_id, archived, status;

-- Company + Status filter counts
INSERT INTO user_filter_counts (user_id, archived, company, status, count)
SELECT 
    user_id, 
    archived, 
    company, 
    status, 
    COUNT(*) as count
FROM job_applications 
GROUP BY user_id, archived, company, status;

-- Populate user_company_counts with existing data
INSERT INTO user_company_counts (user_id, company, archived, count)
SELECT 
    user_id, 
    company, 
    archived, 
    COUNT(*) as count
FROM job_applications 
GROUP BY user_id, company, archived;

-- Update existing stats table with archived counts
UPDATE job_application_stats 
SET 
    total_applications_archived = (
        SELECT COUNT(*) 
        FROM job_applications 
        WHERE user_id = job_application_stats.user_id AND archived = 1
    ),
    total_companies_archived = (
        SELECT COUNT(DISTINCT company) 
        FROM job_applications 
        WHERE user_id = job_application_stats.user_id AND archived = 1
    );