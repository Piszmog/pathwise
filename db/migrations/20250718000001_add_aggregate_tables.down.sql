-- Remove added columns from stats table
ALTER TABLE job_application_stats DROP COLUMN total_applications_archived;
ALTER TABLE job_application_stats DROP COLUMN total_companies_archived;

-- Drop the new tables
DROP TABLE IF EXISTS user_company_counts;
DROP TABLE IF EXISTS user_filter_counts;