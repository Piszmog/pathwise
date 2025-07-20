INSERT INTO users (email, password) VALUES 
  ('test1@example.com', '$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'),
  ('test2@example.com', '$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'),
  ('test3@example.com', '$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm');

INSERT INTO job_application_stats (user_id) 
SELECT id FROM users WHERE email IN ('test1@example.com', 'test2@example.com', 'test3@example.com');

INSERT INTO job_applications (company, title, url, applied_at, user_id) 
SELECT 'Sample Company', 'Sample Title', 'http://sample.com', datetime('now', '-2 days'), id
FROM users WHERE email = 'test3@example.com';

INSERT INTO job_application_status_histories (job_application_id) 
VALUES (last_insert_rowid());
