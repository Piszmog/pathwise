-- Test user with existing job application for specific test scenarios
INSERT INTO users (
	email,
	password
)
VALUES (
	'existing-user@test.com',
	'$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'
);

INSERT INTO job_applications (company, title, url, applied_at, user_id) 
SELECT 
	'Company A', 
	'Title A', 
	'http://companyA/titleA',
	datetime('now', '-2 days'),
	id
FROM users where email = 'existing-user@test.com';

INSERT INTO job_application_status_histories (
	job_application_id
) VALUES (
	last_insert_rowid()
);

INSERT INTO job_application_stats (total_applications, total_companies, total_applied, user_id)
SELECT 1, 1, 1, id FROM users where email = 'existing-user@test.com';
